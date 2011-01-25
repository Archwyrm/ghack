// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "fmt"
    "time"
    "cmpId/cmpId"
    "core/core"
    "pubsub/pubsub"
)

// The core of the game
type Game struct {
    *CmpData
    pubSub chan core.ServiceMsg
}

func (g Game) Id() int      { return cmpId.Game }
func (g Game) Name() string { return "Game" }
func NewGame() *Game        { return &Game{NewCmpData(), make(chan core.ServiceMsg)} }

// TODO: Move functionality to an Init Action
func (g Game) GameLoop() {
    // Initialize stuff
    g.Initialize()

    getState := make(chan State)
    go func() {
        for {
            pos := (<-getState).(Position)
            fmt.Printf("Position is currently: %d, %d\n", pos.X, pos.Y)
        }
    }()

    for {
        g.publish("entities", MsgTick{})
        g.publish("entities", MsgGetState{cmpId.Position, getState})
        time.Sleep(5e9) // 3s
    }
}

// This will happen a lot; might want it to be public
func (g Game) publish(topic string, data interface{}) {
    g.pubSub <- pubsub.PublishMsg{Topic: topic, Data: data}
}

func (g Game) Initialize() {
    entities := NewEntityList()
    g.states[entities.Id()] = entities
    go pubsub.NewPubSub().Run(g.pubSub)

    // The rest is for demo purposes
    player := NewPlayer()
    playerChan := make(chan CmpMsg)
    glueChan := make(chan interface{})
    go func() {
        // This was interesting - directly passing playerChan to pubsub broke, saying
        // cannot use playerChan (type chan CmpMsg) as type chan interface { } in field value
        for {
            cmpMsg, ok := (<-glueChan).(CmpMsg)
            if ok {
                playerChan <- cmpMsg
            }
        }
    }()
    g.pubSub <- pubsub.SubscribeMsg{"entities", glueChan}
    go player.Run(playerChan)
    entities.Entities[player.Name()] = player

    g.publish("entities", MsgAddAction{&Move{1, 1}})
}
