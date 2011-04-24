// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Communications package. Handles all communication with external or remote processes.
package comm

import (
    "net"
    "log"
    "os"
    "io"
    "encoding/binary"
    "bytes"
    "fmt"
    "core"
    "protocol"
    "goprotobuf.googlecode.com/hg/proto"
)

const (
    ProtocolVersion = 1

    lengthBytes = 2                      // Number of bytes to store protobuf length
    maxMsgSize  = 1<<(8*lengthBytes) - 1 // 2^(8 * lengthBytes)
)

var byteOrder = binary.LittleEndian
// Game specific function for creating an avatar, set by game code.
var AvatarFunc func(core.ServiceContext, chan *protocol.Message) (chan core.Msg,
core.UniqueId) = dummyAvatarFunc

// addClient and removeClient are internal messages for manipulating the list
// of clients in a thread safe way
type addClientMsg struct {
    cl *client
}

type removeClientMsg struct {
    cl     *client
    reason string
}

type CommService struct {
    svc     core.ServiceContext
    clients []*client
    address string
}

func NewCommService(svc core.ServiceContext, address string) *CommService {
    return &CommService{svc, make([]*client, 0, 5), address}
}

func (cs *CommService) Run(input chan core.Msg) {
    shutdown := make(chan bool)
    go listen(cs.svc, input, "tcp", cs.address, shutdown)

    cs.svc.Game <- core.MsgTick{input} // Service is ready

    for {
        msg := <-input
        switch m := msg.(type) {
        case addClientMsg:
            cs.clients = append(cs.clients, m.cl)
            log.Println(m.cl.name, "connected")
        case removeClientMsg:
            cs.removeClient(m.cl, m.reason)
        case core.MsgQuit:
            shutdown <- true      // stop listening first so we don't
            cs.removeAllClients() // add any more clients
            return
        case core.MsgTick: // Client state should be updated
            for _, cl := range cs.clients {
                cl.observer <- m
            }
        }
    }
}

func (cs *CommService) removeAllClients() {
    log.Println("Shutting down server")
    for _, cl := range cs.clients {
        cs.removeClient(cl, "")
    }
}

func (cs *CommService) removeClient(cl *client, reason string) {
    found := false
    for i, cur := range cs.clients {
        if cl == cur {
            cs.clients = append(cs.clients[:i], cs.clients[i+1:]...)
            found = true
            break
        }
    }
    if !found {
        return // Client not found, bail
    }

    // TODO: publish disconnection, deal with player entity (when applicable)
    if reason != "" { // Pretty print
        reason = ": " + reason
    }
    log.Println(cl.name, "disconnected"+reason)
    cl.Quit()
}

func listen(svc core.ServiceContext, cs chan<- core.Msg, protocol string,
address string, shutdown chan bool) {
    l, err := net.Listen(protocol, address)
    if err != nil {
        log.Println("Error listening:", err)
        return
    } else {
        log.Println("Server listening on", address)
    }
    defer l.Close()

    accepted := make(chan net.Conn)
    go func() {
        for {
            conn, err := l.Accept()
            if err == os.EINVAL {
                return // socket was closed
            } else if err != nil {
                log.Println("Error accepting connection:", err)
                continue
            }

            accepted <- conn
        }
    }()

    for {
        select {
        case conn := <-accepted:
            go connect(svc, cs, conn)
        case <-shutdown:
            return
        }
    }
}

func connect(svc core.ServiceContext, cs chan<- core.Msg, conn net.Conn) {
    defer logAndClose(conn)

    // Read connect message
    conn.SetReadTimeout(1e9) // 1s
    msg := readMessageOrPanic(conn)
    if msg.Connect == nil {
        panic("Connect message not received!")
    }

    // Check protocol version
    if *msg.Connect.Version != ProtocolVersion {
        // TODO: Send a wrong protocol message, for now just close
        panic(fmt.Sprintf("Wrong protocol version %d, needed %d",
            *msg.Connect.Version, ProtocolVersion))
    }

    // Send connect reply
    msg = makeConnect()
    sendMessageOrPanic(conn, msg)

    // Read login message
    msg = readMessageOrPanic(conn)
    if msg.Login == nil {
        panic("Login message not received!")
    }
    login := msg.Login
    logged_in, reason := startLogin(login)

    // Send login reply
    msg = makeLoginResult(logged_in, reason)
    sendMessageOrPanic(conn, msg)

    cl := newClient(svc, cs, conn, login)
    cs <- addClientMsg{cl}
}

// Recovers from fatal errors, logs them, and closes the connection
func logAndClose(conn net.Conn) {
    if e := recover(); e != nil {
        log.Println(e)
        conn.Close()
    }
}

func sendMessageOrPanic(w io.Writer, msg *protocol.Message) {
    err := sendMessage(w, msg)
    if err != nil {
        panic(err.String())
    }
}

func readMessageOrPanic(r io.Reader) *protocol.Message {
    msg, err := readMessage(r)
    if err != nil {
        panic(err.String())
    }
    return msg
}

func sendMessage(w io.Writer, msg *protocol.Message) os.Error {
    // Marshal protobuf
    bs, err := proto.Marshal(msg)
    if err != nil {
        return err
    }

    // Send pb
    if bs, err = prependByteLength(bs); err != nil {
        return err
    }
    if n, err := w.Write(bs); err != nil {
        return err
    } else if n != len(bs) {
        return os.NewError(fmt.Sprintf("Wrote only %d bytes out of %d bytes!", n, len(bs)))
    }

    return nil
}

func readMessage(r io.Reader) (msg *protocol.Message, err os.Error) {
start:
    // Read length
    length, err := readLength(r)
    if err != nil {
        if err == os.EOF {
            goto start // No data was ready, read again
        } else if err == os.EINVAL {
            return nil, err // Socket closed mid-read
        } else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
            // Socket timed out on read, read again
            goto start
        } else {
            return nil, err
        }
    }

    // Read the message bytes
    bs := make([]byte, length)
    if n, err := io.ReadFull(r, bs); err != nil {
        log.Println("Error reading message bytes:", err)
    } else if n != len(bs) {
        return nil, os.NewError(fmt.Sprintf("Read only %d bytes out of expected %d bytes!", n, length))
    }

    // Unmarshal
    msg = new(protocol.Message)
    if err := proto.Unmarshal(bs, msg); err != nil {
        return nil, err
    }
    return msg, nil
}

// Reads the length of a message
func readLength(r io.Reader) (length uint16, err os.Error) {
    err = binary.Read(r, byteOrder, &length)
    return
}

// Prepends the length of the passed byte array to the array.
// Returns error if byte array is too large.
func prependByteLength(data []byte) ([]byte, os.Error) {
    data_len := len(data)
    if data_len > maxMsgSize {
        return nil, os.NewError("Message size exceeds maxMsgSize")
    }
    length := uint16(data_len)

    buf := new(bytes.Buffer)
    err := binary.Write(buf, byteOrder, length)
    if err != nil {
        return nil, os.NewError(fmt.Sprintf("Binary conversion error: %s", err))
    }
    data = append(buf.Bytes(), data...)
    return data, nil
}

// Initiates a login and returns the result of the login attempt
func startLogin(msg *protocol.Login) (bool, int32) {
    // TODO: Talk to login service
    return true, protocol.LoginResult_ACCEPTED // Logins are still always accepted ;(
}

// Represents remote client. Contains queue of messages to send and permission
// set governing what messages will be accepted and acted upon.
type client struct {
    // Name of client or player name
    name string
    // conn transport to client
    conn net.Conn
    // Permission set mask
    permissions uint32
    // Queue of messages to be sent to client. observer fills this channel.
    SendQueue chan core.Msg
    // Queue of messages received from client. avatar drains this channel.
    // Messages are passed in their original protocol form as we cannot
    // anticipate game defined messages here.
    RecvQueue chan *protocol.Message
    // Msgs meant for observer specifically and *not* the client are sent here.
    // e.g.: tick, quit, etc
    observer chan core.Msg
    // Control channel for avatar
    avatar chan core.Msg
}

// Create a new client and start up send/receive goroutines.
func newClient(svc core.ServiceContext, cs chan<- core.Msg, conn net.Conn,
l *protocol.Login) *client {
    send_ch := make(chan core.Msg)
    recv_ch := make(chan *protocol.Message)
    obs := createObserver(svc, send_ch)
    avatar, uid := AvatarFunc(svc, recv_ch)
    cl := &client{
        name:        *l.Name,
        permissions: proto.GetUint32(l.Permissions),
        conn:        conn,
        SendQueue:   send_ch,
        RecvQueue:   recv_ch,
        observer:    obs,
        avatar:      avatar,
    }
    go cl.RecvLoop(cs)
    go cl.SendLoop(cs)

    // Only attempt to assign control if a real avatar channel was returned,
    // otherwise the uid is just a dummy and should not be sent. This mostly
    // applies to tests.
    if avatar != nil {
        send_ch <- core.MsgAssignControl{uid, false}
    }
    return cl
}

// Receives messages from remote client and acts upon them if appropriate.
func (cl *client) RecvLoop(cs chan<- core.Msg) {
    defer logAndClose(cl.conn)
    for {
        msg, err := readMessage(cl.conn)
        if err != nil {
            // Remove client if something went wrong
            cs <- removeClientMsg{cl, "Reading message from client failed: " + err.String()}
            return
        }
        switch *msg.Type {
        case protocol.Message_Type(protocol.Message_DISCONNECT):
            cs <- removeClientMsg{cl, proto.GetString(msg.Disconnect.ReasonStr)}
            return
        default:
            // TODO: If no proper avatar has been started, this will block, fix?
            // We could check cl.avatar for nil
            cl.RecvQueue <- msg // Forward to avatar
        }
    }
}

// Sends messages over the remote conn that come through the queue.
func (cl *client) SendLoop(cs chan<- core.Msg) {
    defer logAndClose(cl.conn)
    for {
        msg := <-cl.SendQueue
        var err os.Error
        switch m := msg.(type) {
        case MsgAddEntity:
            err = sendMessage(cl.conn, makeAddEntity(int32(m.Uid), m.Name))
        case MsgRemoveEntity:
            err = sendMessage(cl.conn, makeRemoveEntity(int32(m.Uid), m.Name))
        case MsgUpdateState:
            value := packState(m.State)
            err = sendMessage(cl.conn, makeUpdateState(int32(m.Uid), m.State.Name(), value))
        case core.MsgAssignControl:
            err = sendMessage(cl.conn, makeAssignControl(int32(m.Uid), m.Revoked))
        case core.MsgEntityDeath:
            uid, name := m.Entity.Uid, m.Entity.Name
            kuid, kname := m.Killer.Uid, m.Killer.Name
            err = sendMessage(cl.conn, makeEntityDeath(int32(uid), name, int32(kuid), kname))
        case core.MsgCombatHit:
            auid, aname := m.Attacker.Uid, m.Attacker.Name
            vuid, vname := m.Victim.Uid, m.Victim.Name
            err = sendMessage(cl.conn, makeCombatHit(int32(auid), aname, int32(vuid), vname, m.Damage))
        }
        // Remove client if something went wrong
        if err != nil {
            cs <- removeClientMsg{cl, "Sending message to client failed: " + err.String()}
            return
        }
    }
}

// Disconnects client and closes all client resources.
func (cl *client) Quit() {
    cl.conn.Close()

    // Close this client's observer and avatar
    quit := core.MsgQuit{}
    cl.observer <- quit
    cl.avatar <- quit
}

// Return default values to satisfy tests, if returned chan is used, will cause
// panic (of course)
func dummyAvatarFunc(core.ServiceContext, chan *protocol.Message) (chan core.Msg,
core.UniqueId) {
    return nil, 1
}
