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
func MsgBuffer(ch chan core.Msg) chan core.Msg {
    in := make(chan core.Msg)
    go func() {
        buf := make([]core.Msg, 0, 2)
        var msg core.Msg
        var out chan core.Msg // Start as nil so we don't send
        for {
            select {
            case next := <-in: // Read next message
                out = ch // Enable send
                if msg != nil {
                    buf = append(buf, next) // Save in queue
                } else if len(buf) > 0 {
                    msg = buf[0] // Buf has messages, pop the next in queue
                    buf = buf[1:len(buf)]
                } else {
                    msg = next
                }
            case out <- msg: // Try to send current msg
                msg = nil // Mark msg as sent
                if len(buf) == 0 {
                    out = nil // Disable send
                }
            }
        }
    }()
    return in
}
