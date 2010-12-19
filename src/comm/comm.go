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
    conn.SetReadTimeout(1e9) // 1s
    length, err := readLength(conn)
    if err != nil {
        log.Println("Error reading message length:", err)
        conn.Close()
        return
    }

    // Read the message bytes
    bs := make([]byte, length)
    if _, err := io.ReadFull(conn, bs); err != nil {
        log.Println("Error reading message:", err)
    }

    // Unmarshal connect pb
    connect := new(protocol.Connect)
    if err := proto.Unmarshal(bs, connect); err != nil {
        log.Println("Error unmarshaling connect msg:", err)
        conn.Close()
        return
    }

    // TODO: Send a wrong protocol message, for now just close
    vstr := fmt.Sprintf("%d", ProtocolVersion)
    if *connect.Version != vstr {
        log.Println("Wrong protocol version", *connect.Version, "need", vstr)
        conn.Close()
        return
    }

    // Marshal connect reply pb
    connect.Version = &vstr
    bs, err = proto.Marshal(connect)
    if err != nil {
        log.Println("Error marshaling version reply:", err)
        conn.Close()
        return
    }

    // Send pb
    if bs, err = PrependByteLength(bs); err != nil {
        log.Println("Error:", err)
        conn.Close()
        return
    }
    if _, err := conn.Write(bs); err != nil {
        log.Println("Error sending version reply:", err)
        conn.Close()
        return
    }

    conn.Close()
    // TODO: Wait for avatar request
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
