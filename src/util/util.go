// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Utility functions for general use
package util

import (
    "core"
)

// Drains and discards messages from the passed channel indefinitely
func Drain(ch chan core.Msg) {
    for {
        <-ch
    }
}

// Drains and discards messages from the passed channel until a MsgQuit is received
func DrainUntilQuit(ch chan core.Msg) {
    for {
        if _, ok := (<-ch).(core.MsgQuit); ok {
            return
        }
    }
}

// Variable length buffer for the passed channel. Returns a channel for input.
// TODO: Write test to ensure that this function always does the right thing
func MsgBuffer(ch chan core.Msg) chan core.Msg {
    in := make(chan core.Msg)
    go func() {
        buf := make([]core.Msg, 0, 2)
        var msg core.Msg
        var out chan core.Msg // Start as nil so we don't send

        // Alternate between receiving from in and sending on out. Each
        // received message gets appended onto buf and then the first value is
        // popped off and sent.
        for {
            if len(buf) == 0 {
                out = nil // Disable send
            } else {
                out = ch              // Enable send
                msg = buf[0]          // Set value to send
                buf = buf[1:len(buf)] // Discard the value
            }
            select {
            case next := <-in: // Read next message
                buf = append(buf, next) // Save in queue
            case out <- msg: // Send message, if enabled
            }
        }
    }()
    return in
}

// Performs asynchronus send of msg to ch. Initially tries to send directly to
// the channel, if this is not immediately possible, a goroutine is started to
// perform the send. This has the effect of Send not blocking.
func Send(ch chan core.Msg, msg core.Msg) {
    select {
    case ch <- msg:
    default:
        go func(ch chan core.Msg, data interface{}) {
            ch <- data
        }(ch, msg)
    }
}
