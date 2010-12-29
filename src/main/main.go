// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Server main.
package main

import (
    "fmt"
)

func main() {
    fmt.Printf("Game started\n")

    game := NewGame()
    game.GameLoop()

    fmt.Printf("Exiting\n")
}
