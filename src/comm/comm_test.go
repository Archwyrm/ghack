// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "testing"
    "net"
    "time"
    "protocol"
    "pubsub"
    "core"
    "util"
)

// Starts the server with a default ServiceContext for tests that don't need it
func startServer(t *testing.T) (svc *CommService, cs chan core.Msg) {
    return startServerWithCtx(t, core.NewServiceContext())
}

// Starts the server with a user specificied ServiceContext
func startServerWithCtx(t *testing.T,
ctx core.ServiceContext) (svc *CommService, cs chan core.Msg) {
    // Start new service on port 9190
    svc = NewCommService(ctx, ":9190")
    go util.Drain(ctx.Game) // For service ready msg
    cs = ctx.Comm
    go svc.Run(cs)

    // Start game and pubsub so observers don't lock up
    go core.NewGame(ctx).Run(ctx.Game)
    go pubsub.NewPubSub(ctx).Run(ctx.PubSub)

    // Give time for the service to start listening
    time.Sleep(1e8) // 100 ms
    return svc, cs
}

func newTestClient(t *testing.T) (fd net.Conn) {
    // Connect to the listening service and send
    fd, err := net.Dial("tcp", "localhost:9190")
    if err != nil {
        t.Fatalf("Could not connect to comm:", err)
    }

    // Wait 1s to read a reply
    fd.SetReadTimeout(1e9) // 1s
    return
}

// Do the connection handshake
func connectClient(t *testing.T, fd net.Conn) {
    // Recover from any panics sent by {send,read}Msg() and print the error
    failure := "Error sending connect message"
    defer func() {
        if e := recover(); e != nil {
            t.Fatalf("%s: %v", failure, e)
            fd.Close()
        }
    }()

    // Create protocol buffer to initiate connection
    connect := makeConnect()
    sendMessageOrPanic(fd, connect)

    failure = "Connect message not received"
    msg := readMessageOrPanic(fd)
    if msg.Connect == nil {
        t.Fatalf(failure)
    }
    reply_pb := msg.Connect

    // Since the client and server are running the same code, the version
    // strings should be exact
    if *reply_pb.Version != *connect.Connect.Version {
        t.Error("Version strings do not match")
    }

    // Send login message
    failure = "Error sending login message"
    login := makeLogin("TestPlayer", "passwordHash", 0)
    sendMessageOrPanic(fd, login)

    // Read login result message
    failure = "Login result message not received"
    msg = readMessageOrPanic(fd)
    if msg.LoginResult == nil {
        t.Fatalf(failure)
    }
    result := msg.LoginResult
    if *result.Succeeded != true {
        t.Fatalf("Login failed!")
    }
}

// Tests the connection handshake
func TestConnect(t *testing.T) {
    _, cs := startServer(t)
    fd := newTestClient(t)

    // Do connect/login handshake
    connectClient(t, fd)

    // Send disconnect
    failure := "Error sending disconnect"
    serviceClosed := false
    defer func() {
        if e := recover(); e != nil {
            t.Fatalf("%s: %v", failure, e)
            fd.Close()
        }

        if !serviceClosed {
            cs <- core.MsgQuit{}
        }
    }()
    disconnect := makeDisconnect(protocol.Disconnect_QUIT, "Test finished")
    sendMessageOrPanic(fd, disconnect)

    fd.Close()

    time.Sleep(1e7) // 10 ms, give time for disconnect to process
}

func TestServerQuit(t *testing.T) {
    svc, cs := startServer(t)

    fd := newTestClient(t)
    connectClient(t, fd)

    cs <- core.MsgQuit{}
    time.Sleep(1e7) // Block thread 10ms to let the server respond
    if len(svc.clients) > 0 {
        t.Fatalf("Client not removed from server list")
    }

    errMsg := make(chan string)
    go func() {
        if _, err := fd.Read(make([]byte, 10)); err == nil {
            errMsg <- "Client connection not closed"
        } else {
            errMsg <- ""
        }
    }()
    go func() {
        // if the socket is open, the Read might block
        time.Sleep(1e8) // 100ms
        errMsg <- "Client connection not closed after 100ms"
    }()
    if err := <-errMsg; err != "" {
        t.Fatalf(err)
    }

    go func() {
        _, err := net.Dial("tcp", "localhost:9190")
        if err == nil {
            t.Fatalf("Server didn't shut down")
        }
    }()

    time.Sleep(1e8) // Wait 100ms to make sure we can't connect
}
