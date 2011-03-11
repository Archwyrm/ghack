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
    svc core.ServiceContext
    *core.CmpData
    nextUid core.UniqueId
}

func (g Game) Id() core.EntityId { return cmpId.Game }
func (g Game) Name() string      { return "Game" }
func NewGame(svc core.ServiceContext) *Game {
    var uid core.UniqueId = 0 // Game uid is always zero
    return &Game{svc, core.NewCmpData(uid), uid + 1}
}

func (g *Game) GameLoop() {
    // Initialize stuff
    entities := NewEntityList()
    g.SetState(entities)

    spider := NewSpider(g.getUid())
    spiderChan := make(chan core.Msg)
    entities.Entities[spiderChan] = spider
    go spider.Run(spiderChan)

    msg := core.MsgAddAction{&Move{&s3dm.V3{1, 1, 1}}}
    spiderChan <- msg

    reply := make(chan core.State)
    msg2 := core.MsgGetState{cmpId.Position, reply}
    tick_msg := core.MsgTick{g.svc.Game}

    // List of up to date entities
    updated := make(map[chan core.Msg]bool, len(entities.Entities))

    for {
        // Tell all the entities that a new tick has started
        ent_num := len(entities.Entities) // Store ent count for *this* tick
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
                if len(updated) == ent_num {
                    // Clear out list, use current number of entities for next tick
                    updated = make(map[chan core.Msg]bool, len(entities.Entities))
                    goto update_end
                }
            case core.MsgListEntities:
                m.Reply <- g.makeEntityList()
            case core.MsgSpawnPlayer:
                g.spawnPlayer(m)
            }
        }
    update_end:
        g.svc.Comm <- core.MsgTick{g.svc.Game}

        // Just a little output for debugging
        spiderChan <- msg2
        pos := (<-reply).(Position)
        fmt.Printf("Position is currently: %g, %g\n", pos.Position.X, pos.Position.Y)
        time.Sleep(3e9) // 3s
    }
}

func (g *Game) makeEntityList() core.MsgListEntities {
    list := g.GetState(cmpId.EntityList).(EntityList).Entities
    ents := make([]*core.EntityDesc, len(list))

    i := 0
    for ch, ent := range list { // Fill the len(list) slots in ents
        ents[i] = core.NewEntityDesc(ent, ch)
        i++
    }
    return core.MsgListEntities{nil, ents}
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
    list := g.GetState(cmpId.EntityList).(EntityList).Entities
    p := NewPlayer(g.getUid())
    ch := make(chan core.Msg)
    list[ch] = p
    go p.Run(ch)
    // TODO: Send entity channel when there is something listening on the other end of Reply..
    //msg.Reply <- ch
    log.Printf("%s joined the game", msg.Name)
}
