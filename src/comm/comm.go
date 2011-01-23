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
    "core/core"
    "protocol/protocol"
    "goprotobuf.googlecode.com/hg/proto"
)

const (
    ProtocolVersion = 1

    lengthBytes = 2                      // Number of bytes to store protobuf length
    maxMsgSize  = 1<<(8*lengthBytes) - 1 // 2^(8 * lengthBytes)
)

var byteOrder = binary.LittleEndian

// addClient and removeClient are internal messages for manipulating the list
// of clients in a thread safe way
type addClientMsg struct {
    cl *client
}

func (x addClientMsg) Name() string { return "addClientMsg" }

type removeClientMsg struct {
    cl     *client
    reason string
}

func (x removeClientMsg) Name() string { return "addClientMsg" }

// Message to shut the server down gracefully
type ShutdownServerMsg struct{}

func (msg ShutdownServerMsg) Name() string { return "ShutdownServerMsg" }

type CommService struct {
    clients []*client
    address string
}

func NewCommService(address string) *CommService {
    return &CommService{make([]*client, 5), address}
}

func (cs *CommService) Run(input chan core.ServiceMsg) {
    shutdown := make(chan bool)
    go listen(input, "tcp", cs.address, shutdown)

    for {
        msg := <-input
        switch m := msg.(type) {
        case addClientMsg:
            cs.clients = append(cs.clients, m.cl)
            log.Println(m.cl.name, "connected")
        case removeClientMsg:
            cs.removeClient(m)
        case ShutdownServerMsg:
            shutdown <- true      // stop listening first so we don't
            cs.removeAllClients() // add any more clients
            return
        }
    }
}

func (cs *CommService) removeAllClients() {
    log.Println("Shutting down server")
    num_clients := len(cs.clients)
    for _, cl := range cs.clients {
        if cl == nil {
            continue
        }

        cl.conn.Close()
    }

    cs.clients = make([]*client, num_clients)
}

func (cs *CommService) removeClient(msg removeClientMsg) {
    for i, cur := range cs.clients {
        if msg.cl == cur {
            cs.clients = append(cs.clients[:i], cs.clients[i+1:]...)
            break
        }
    }
    // TODO: publish disconnection, deal with player entity (when applicable)
    if msg.reason != "" { // Pretty print
        msg.reason = ": " + msg.reason
    }
    log.Println(msg.cl.name, "disconnected"+msg.reason)
}

func listen(cs chan<- core.ServiceMsg, protocol string, address string, shutdown chan bool) {
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
            go connect(cs, conn)
        case <-shutdown:
            return
        }
    }
}

func connect(cs chan<- core.ServiceMsg, conn net.Conn) {
    defer logAndClose(conn)

    // Read connect message
    conn.SetReadTimeout(1e9) // 1s
    msg, ok := readMessage(conn)
    if !ok || msg.Connect == nil {
        panic("Connect message not received!")
    }
    connectMsg := msg.Connect

    // Check protocol version
    if *connectMsg.Version != ProtocolVersion {
        // TODO: Send a wrong protocol message, for now just close
        panic(fmt.Sprintf("Wrong protocol version %d, needed %d",
            *connectMsg.Version, ProtocolVersion))
    }

    // Send connect reply
    connectMsg.Version = proto.Uint32(ProtocolVersion)
    msg = &protocol.Message{Connect: connectMsg,
        Type: protocol.NewMessage_Type(protocol.Message_CONNECT)}
    sendMessage(conn, msg)

    // Read login message
    msg, ok = readMessage(conn)
    if !ok || msg.Login == nil {
        panic("Login message not received!")
    }
    login := msg.Login
    // TODO: Handle proper login here

    // Send login reply
    result := &protocol.LoginResult{Succeeded: proto.Bool(true)}
    msg = &protocol.Message{LoginResult: result,
        Type: protocol.NewMessage_Type(protocol.Message_LOGINRESULT)}
    sendMessage(conn, msg)

    cs <- addClientMsg{newClient(cs, conn, login)}
}

// Recovers from fatal errors, logs them, and closes the connection
func logAndClose(conn net.Conn) {
    if e := recover(); e != nil {
        log.Println(e)
        conn.Close()
    }
}

func sendMessage(w io.Writer, msg *protocol.Message) {
    // Marshal protobuf
    bs, err := proto.Marshal(msg)
    if err != nil {
        panic("Error marshaling message: " + err.String())
    }

    // Send pb
    if bs, err = prependByteLength(bs); err != nil {
        panic("Cannot prepend: " + err.String())
    }
    if _, err = w.Write(bs); err != nil {
        panic("Error writing message: " + err.String())
    }
}

func readMessage(r io.Reader) (msg *protocol.Message, ok bool) {
start:
    // Read length
    length, err := readLength(r)
    if err != nil {
        if err == os.EOF {
            goto start // No data was ready, read again
        } else if err == os.EINVAL {
            return nil, false // Socket closed mid-read
        }
        panic("Error reading message length: " + err.String())
    }

    // Read the message bytes
    bs := make([]byte, length)
    if _, err := io.ReadFull(r, bs); err != nil {
        log.Println("Error reading message bytes:", err)
    }

    // Unmarshal
    msg = new(protocol.Message)
    if err := proto.Unmarshal(bs, msg); err != nil {
        panic("Error unmarshaling msg: " + err.String())
    }
    return msg, true
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

// Represents remote client. Contains queue of messages to send and permission
// set governing what messages will be accepted and acted upon.
type client struct {
    name        string
    conn        net.Conn
    SendQueue   chan core.ServiceMsg
    permissions uint32
}

// Create a new client and start up send/receive goroutines.
func newClient(cs chan<- core.ServiceMsg, conn net.Conn, l *protocol.Login) *client {
    ch := make(chan core.ServiceMsg)
    cl := &client{*l.Name, conn, ch, proto.GetUint32(l.Permissions)}
    go cl.RecvLoop(cs)
    go cl.SendLoop()
    return cl
}

// Receives messages from remote client and acts upon them if appropriate.
func (cl *client) RecvLoop(cs chan<- core.ServiceMsg) {
    defer logAndClose(cl.conn)
    for {
        msg, ok := readMessage(cl.conn)
        if !ok {
            cs <- removeClientMsg{cl, "Client hung up unexpectedly"}
            return
        }
        switch *msg.Type {
        case protocol.Message_Type(protocol.Message_DISCONNECT):
            cs <- removeClientMsg{cl, proto.GetString(msg.Disconnect.ReasonStr)}
            return
        }
    }
}

// Sends messages over the remote conn that come through the queue.
func (cl *client) SendLoop() {
    defer logAndClose(cl.conn)
    for {
        msg := <-cl.SendQueue
        if msg == nil && closed(cl.SendQueue) {
            return
        }
        switch msg.(type) {
        // TODO: Handle messages in the send queue
        }
    }
}
