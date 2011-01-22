// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "fmt"
    "time"
    "core/core"
    "cmpId/cmpId"
)

// The core of the game
type Game struct {
    *core.CmpData
}

func (g Game) Id() int      { return cmpId.Game }
func (g Game) Name() string { return "Game" }
func NewGame() *Game        { return &Game{core.NewCmpData()} }

// TODO: Move functionality to an Init Action
func (g Game) GameLoop() {
    // Initialize stuff
    entities := NewEntityList()
    g.SetState(entities)

    player := NewPlayer()
    entities.Entities[player.Name()] = player
    playerChan := make(chan core.CmpMsg)
    go player.Run(playerChan)

    msg := core.MsgAddAction{&Move{1, 1}}
    playerChan <- msg

    reply := make(chan core.State)
    msg2 := core.MsgGetState{cmpId.Position, reply}
    tick_msg := core.MsgTick{}
    playerChan <- tick_msg // Tick once

    for {
        playerChan <- msg2
        pos := (<-reply).(Position)
        fmt.Printf("Position is currently: %d, %d\n", pos.X, pos.Y)
        playerChan <- tick_msg
        time.Sleep(5e9) // 3s
    }
}
