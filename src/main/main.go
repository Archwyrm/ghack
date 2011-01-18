// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Server main.
package main

import (
    "fmt"
    "core/core"
    "comm/comm"
)

func main() {
    fmt.Printf("Game started\n")

    svc := comm.NewCommService(":9190")
    go svc.Run(make(chan core.ServiceMsg))

    game := NewGame()
    game.GameLoop()

    fmt.Printf("Exiting\n")
}
