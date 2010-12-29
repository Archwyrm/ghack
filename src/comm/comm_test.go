// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm_test

import (
    "testing"
    "net"
    "fmt"
    "time"
    "encoding/binary"
    "comm/comm"
    "protocol/protocol"
    "core/core"
    "goprotobuf.googlecode.com/hg/proto"
)

// Tests the connection handshake
func TestConnect(t *testing.T) {
    // Start new service on port 9190
    svc := comm.NewCommService(":9190")
    cs := make(chan core.ServiceMsg)
    go svc.Run(cs)
    // Give time for the service to start listening
    time.Sleep(1e8) // 100 ms

    vstr := fmt.Sprintf("%d", comm.ProtocolVersion)
    // Create protocol buffer to initiate connection
    connect := &protocol.Connect{&vstr, nil}
    msg := &protocol.Message{Connect: connect,
        Type: protocol.NewMessage_Type(protocol.Message_CONNECT)}
    data, err := proto.Marshal(msg)
    if err != nil {
        t.Fatalf("Marshaling error: %s", err)
    }
    data, err = comm.PrependByteLength(data)
    if err != nil {
        t.Fatalf("Error: %s", err)
    }

    // Connect to the listening service and send
    fd, err := net.Dial("tcp", "", "localhost:9190")
    if err != nil {
        t.Fatalf("Could not connect to comm:", err)
    }
    if _, err := fd.Write(data); err != nil {
        t.Fatalf("Error writing connect message:", err)
    }

    // Wait 1s to read a reply
    fd.SetReadTimeout(1e9) // 1s
    // Read the message length
    bs := make([]byte, 2)
    if _, err := fd.Read(bs); err != nil {
        t.Fatalf("Error reading socket: %s", err)
    }
    length := binary.LittleEndian.Uint16(bs)

    // Read the message
    reply := make([]byte, length)
    if _, err := fd.Read(reply); err != nil {
        t.Fatalf("Error reading socket: %s", err)
    }

    // Unmarshal the received data
    err2 := proto.Unmarshal(reply, msg)
    if err2 != nil {
        t.Fatalf("Unmarshaling error: %s", err2)
    }
    reply_pb := msg.Connect
    if reply_pb == nil {
        t.Fatalf("Connect message not received!")
    }

    // Since the client and server are running the same code, the version
    // strings should be exact
    if *reply_pb.Version != *connect.Version {
        t.Error("Version strings do not match!")
    }
}
