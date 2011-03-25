// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "fmt"
    "log"
    "time"
    "github.com/tm1rbrt/s3dm"
    "core"
    "cmpId"
)

// The core of the game
type Game struct {
    svc     core.ServiceContext
    ents    map[chan core.Msg]core.Entity
    nextUid core.UniqueId
}

func NewGame(svc core.ServiceContext) *Game {
    var uid core.UniqueId = 0 // Game uid is always zero
    ents := make(map[chan core.Msg]core.Entity)
    return &Game{svc, ents, uid + 1}
}

func (g *Game) Run(input chan core.Msg) {
    // Initialize stuff

    spider := NewSpider(g.getUid())
    spiderChan := make(chan core.Msg)
    g.ents[spiderChan] = spider
    go spider.Run(spiderChan)

    msg := core.MsgAddAction{&Move{&s3dm.V3{1, 1, 1}}}
    spiderChan <- msg

    reply := make(chan core.State)
    msg2 := core.MsgGetState{cmpId.Position, reply}
    tick_msg := core.MsgTick{input}

    // List of up to date entities
    updated := make(map[chan core.Msg]bool, len(g.ents))

    for {
        // Tell all the entities that a new tick has started
        ent_num := len(g.ents) // Store ent count for *this* tick
        for ent := range g.ents {
            ent <- tick_msg
        }

        // Listen for any service messages
        // Break out of loop once all entities have updated
        for {
            msg := <-input
            switch m := msg.(type) {
            case core.MsgTick:
                updated[m.Origin] = true // bool value doesn't matter
                if len(updated) == ent_num {
                    // Clear out list, use current number of entities for next tick
                    updated = make(map[chan core.Msg]bool, len(g.ents))
                    goto update_end
                }
            case core.MsgListEntities:
                m.Reply <- g.makeEntityList()
            case core.MsgSpawnPlayer:
                g.spawnPlayer(m)
            }
        }
    update_end:
        g.svc.Comm <- tick_msg

        // Just a little output for debugging
        spiderChan <- msg2
        pos := (<-reply).(Position)
        fmt.Printf("Position is currently: %g, %g\n", pos.Position.X, pos.Position.Y)
        time.Sleep(3e9) // 3s
    }
}

func (g *Game) makeEntityList() core.MsgListEntities {
    list := make([]*core.EntityDesc, len(g.ents))

    i := 0
    for ch, ent := range g.ents { // Fill the len(g.ents) slots in list
        list[i] = core.NewEntityDesc(ent, ch)
        i++
    }
    return core.MsgListEntities{nil, list}
}

// Returns the next available unique id
func (g *Game) getUid() core.UniqueId {
    uid := g.nextUid
    g.nextUid++
    return uid
}

// Starting to get somewhat game specific?

// Creates a new player entity for a requesting client
func (g *Game) spawnPlayer(msg core.MsgSpawnPlayer) {
    // Lines like the following are rather unwieldy.. Somewhat of an argument for
    // making game a service rather than an entity
    p := NewPlayer(g.getUid())
    ch := make(chan core.Msg)
    g.ents[ch] = p
    go p.Run(ch)
    go func(uid core.UniqueId) {
        msg.Reply <- core.MsgAssignControl{uid, false}
    }(p.Uid())
    log.Printf("%s joined the game", msg.Name)
}
