// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "fmt"
    "time"
    "github.com/tm1rbrt/s3dm"
    "core/core"
    "cmpId/cmpId"
)

// The core of the game
type Game struct {
    svc core.ServiceContext
    *core.CmpData
}

func (g Game) Id() core.EntityId { return cmpId.Game }
func (g Game) Name() string      { return "Game" }
func NewGame(svc core.ServiceContext) *Game {
    return &Game{svc, core.NewCmpData()}
}

func (g *Game) GameLoop() {
    // Initialize stuff
    entities := NewEntityList()
    g.SetState(entities)

    player := NewPlayer()
    playerChan := make(chan core.Msg)
    entities.Entities[playerChan] = player
    go player.Run(playerChan)

    msg := core.MsgAddAction{&Move{&s3dm.V3{1, 1, 1}}}
    playerChan <- msg

    reply := make(chan core.State)
    msg2 := core.MsgGetState{cmpId.Position, reply}
    tick_msg := core.MsgTick{g.svc.Game}

    // List of up to date entities
    updated := make(map[chan core.Msg]bool, len(entities.Entities))

    for {
        // Tell all the entities that a new tick has started
        for ent := range entities.Entities {
            ent <- tick_msg
        }

        // Listen for any service messages
        // Break out of loop once all entities have updated
        for {
            msg := <-g.svc.Game
            switch m := msg.(type) {
            case core.MsgTick:
                updated[m.Origin] = true // bool value doesn't matter
                if len(updated) == len(entities.Entities) {
                    // Clear out list
                    updated = make(map[chan core.Msg]bool, len(entities.Entities))
                    goto update_end
                }
            case core.MsgListEntities:
                m.Reply <- g.makeEntityList()
            }
        }
    update_end:
        g.svc.Comm <- core.MsgTick{g.svc.Game}

        // Just a little output for debugging
        playerChan <- msg2
        pos := (<-reply).(Position)
        fmt.Printf("Position is currently: %g, %g\n", pos.Position.X, pos.Position.Y)
        time.Sleep(3e9) // 3s
    }
}

func (g *Game) makeEntityList() core.MsgListEntities {
    list := g.GetState(cmpId.EntityList).(EntityList)
    length := len(list.Entities)
    chans := make([]chan core.Msg, 0, length)
    ids := make([]core.EntityId, 0, length)
    names := make([]string, 0, length)

    for ch, ent := range list.Entities {
        chans = append(chans, ch)
        ids = append(ids, 0) // TODO: Send proper unique id
        names = append(names, ent.Name())
    }
    return core.MsgListEntities{nil, chans, ids, names}
}
