// Copyright 2010 The ghack Authors. All rights reserved.
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

type CommService struct {
    address string
}

func NewCommService(address string) *CommService {
    return &CommService{address}
}

func (cs *CommService) Run(input chan core.ServiceMsg) {
    go listen("tcp", cs.address)

    for {
        <-input
    }
}

func listen(protocol string, address string) {
    l, err := net.Listen("tcp", address)
    defer l.Close()
    if err != nil {
        log.Println("Error listening:", err)
    } else {
        log.Println("Server listening on", address)
    }

    for { // TODO: Need to be able to shutdown the server remotely
        conn, err := l.Accept()
        if err != nil {
            log.Println("Error accepting connection:", err)
            continue
        }
        go connect(conn)
    }
}

func connect(conn net.Conn) {
    // Recover from fatal errors by closing the connection
    // This goroutine then exits immediately afterwards
    defer func() {
        if e := recover(); e != nil {
            log.Println(e)
            conn.Close()
        }
    }()

    // Read connect message
    conn.SetReadTimeout(1e9) // 1s
    msg := readMessage(conn)
    connect := msg.Connect
    if connect == nil {
        panic("Connect message not received!")
    }

    // Check protocol version
    vstr := fmt.Sprintf("%d", ProtocolVersion)
    if *connect.Version != vstr {
        // TODO: Send a wrong protocol message, for now just close
        panic(fmt.Sprintf("Wrong protocol version %s, needed %s",
            *connect.Version, vstr))
    }

    // Send connect reply
    connect.Version = &vstr
    msg = &protocol.Message{Connect: connect,
        Type: protocol.NewMessage_Type(protocol.Message_CONNECT)}
    sendMessage(conn, msg)
}

func sendMessage(w io.Writer, msg *protocol.Message) {
    // Marshal protobuf
    bs, err := proto.Marshal(msg)
    if err != nil {
        panic("Error marshaling message:" + err.String())
    }

    // Send pb
    if bs, err = PrependByteLength(bs); err != nil {
        panic("Cannot prepend:" + err.String())
    }
    if _, err = w.Write(bs); err != nil {
        panic("Error writing message:" + err.String())
    }
}

func readMessage(r io.Reader) (msg *protocol.Message) {
    // Read length
    length, err := readLength(r)
    if err != nil {
        panic("Error reading message length:" + err.String())
    }

    // Read the message bytes
    bs := make([]byte, length)
    if _, err := io.ReadFull(r, bs); err != nil {
        log.Println("Error reading message bytes:", err)
    }

    // Unmarshal
    msg = new(protocol.Message)
    if err := proto.Unmarshal(bs, msg); err != nil {
        panic("Error unmarshaling msg:" + err.String())
    }
    return msg
}

// Reads the length of a message
func readLength(r io.Reader) (length uint16, err os.Error) {
    err = binary.Read(r, byteOrder, &length)
    return
}

// Prepends the length of the passed byte array to the array.
// Returns error if byte array is too large.
// Public for testing purposes only.
func PrependByteLength(data []byte) ([]byte, os.Error) {
    data_len := len(data)
    if lengthBytes > maxMsgSize {
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
