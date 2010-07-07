package main

import (
    "fmt"
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
    g.states[entities.Name()] = entities

    player := NewPlayer()
    entities.Entities[player.Name()] = player
    playerChan := make(chan CmpMsg)
    go player.Run(playerChan)

    msg := CmpMsg{Id: MsgAddAction, Action: &Move{1, 1}}
    playerChan <- msg

    reply := make(chan State)
    msg2 := CmpMsg{Id: MsgGetState, StateId: "Position", StateReply: reply}
    tick_msg := CmpMsg{Id: MsgTick}
    playerChan <- tick_msg // Tick once

    for i := 0; i < 3; i++ {
        playerChan <- msg2
        pos := (<-reply).(Position)
        fmt.Printf("Position is currently: %d, %d\n", pos.X, pos.Y)
        playerChan <- tick_msg
    }
}
