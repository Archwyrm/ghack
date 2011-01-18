// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "fmt"
    "time"
    "cmpId/cmpId"
)

// The core of the game
type Game struct {
    *CmpData
}

func (g Game) Id() int      { return cmpId.Game }
func (g Game) Name() string { return "Game" }
func NewGame() *Game        { return &Game{NewCmpData()} }

// TODO: Move functionality to an Init Action
func (g Game) GameLoop() {
    // Initialize stuff
    entities := NewEntityList()
    g.states[entities.Id()] = entities

    player := NewPlayer()
    entities.Entities[player.Name()] = player
    playerChan := make(chan CmpMsg)
    go player.Run(playerChan)

    msg := MsgAddAction{&Move{1, 1}}
    playerChan <- msg

    reply := make(chan State)
    msg2 := MsgGetState{cmpId.Position, reply}
    tick_msg := MsgTick{}
    playerChan <- tick_msg // Tick once

    for {
        playerChan <- msg2
        pos := (<-reply).(Position)
        fmt.Printf("Position is currently: %d, %d\n", pos.X, pos.Y)
        playerChan <- tick_msg
        time.Sleep(5e9) // 3s
    }
}
