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
