// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "testing"
    "net"
    "time"
    "os"
    "protocol/protocol"
    "core/core"
    "goprotobuf.googlecode.com/hg/proto"
)

// Tests the connection handshake
func TestConnect(t *testing.T) {
    // Start new service on port 9190
    svc := NewCommService(":9190")
    cs := make(chan core.ServiceMsg)
    go svc.Run(cs)
    // Give time for the service to start listening
    time.Sleep(1e8) // 100 ms

    // Create protocol buffer to initiate connection
    connect := &protocol.Connect{proto.Uint32(ProtocolVersion), nil, nil}
    msg := &protocol.Message{Connect: connect,
        Type: protocol.NewMessage_Type(protocol.Message_CONNECT)}

    // Connect to the listening service and send
    fd, err := net.Dial("tcp", "", "localhost:9190")
    if err != nil {
        t.Fatalf("Could not connect to comm:", err)
    }
    // Recover from any panics from sending/receiving and print the error
    failure := "Error sending connect message:"
    defer func() {
        if e := recover(); e != nil {
            if err, ok := e.(os.Error); ok {
                t.Fatalf(failure + err.String())
            } else {
                t.Fatalf(failure)
            }
            fd.Close()
        }
    }()
    sendMessage(fd, msg)

    // Wait 1s to read a reply
    fd.SetReadTimeout(1e9) // 1s
    failure = "Connect message not received:"
    msg = readMessage(fd)
    reply_pb := msg.Connect
    if reply_pb == nil {
        t.Fatalf("Connect message not received!")
    }

    // Since the client and server are running the same code, the version
    // strings should be exact
    if *reply_pb.Version != *connect.Version {
        t.Error("Version strings do not match!")
    }

    // Send login message
    login := &protocol.Login{Name: proto.String("TestPlayer")}
    msg = &protocol.Message{Login: login,
        Type: protocol.NewMessage_Type(protocol.Message_LOGIN)}
    failure = "Error sending login message:"
    sendMessage(fd, msg)

    // Read login result message
    failure = "Login result message not received:"
    msg = readMessage(fd)
    result := msg.LoginResult
    if result == nil {
        t.Fatalf("Login result message not received!")
    }
    if *result.Succeeded != true {
        t.Fatalf("Login failed!")
    }

    // Send disconnect
    failure = "Sending disconnect:"
    disconn := &protocol.Disconnect{protocol.NewDisconnect_Reason(protocol.Disconnect_QUIT),
        proto.String("Test finished"), nil}
    msg = &protocol.Message{Disconnect: disconn,
        Type: protocol.NewMessage_Type(protocol.Message_DISCONNECT)}
    sendMessage(fd, msg)

    fd.Close()
    time.Sleep(1e6) // 1 ms, give time for disconnect to process
}
