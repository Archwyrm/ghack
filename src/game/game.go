// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package game

import (
    "reflect"
    "time"
    "core"
    "pubsub"
)

var tick_rate int64 = 60            // Ticks per second
var skip_ns int64 = 1e9 / tick_rate // Nanosecond interval per tick

// Function that initializes the game state.
var InitFunc func(g *Game, svc core.ServiceContext)

// Manages game data and runs the main loop.
type Game struct {
    svc     core.ServiceContext
    ents    map[chan core.Msg]core.Entity
    nextUid core.UniqueId
    // This function is used to spawn new players when a client joins. As this
    // service has no concept of any specific entities it must be set by the specific game.
    PlayerFunc func(core.UniqueId) core.Entity
}

func NewGame(svc core.ServiceContext) *Game {
    var uid core.UniqueId = 0 // Game uid is always zero
    ents := make(map[chan core.Msg]core.Entity)
    return &Game{svc, ents, uid + 1, nil}
}

func (g *Game) Run(input chan core.Msg) {
    g.waitOnServiceStart(input)
    InitFunc(g, g.svc)

    // List of up to date entities
    updated := make(map[chan core.Msg]bool, len(g.ents))
    tick_msg := core.MsgTick{input}

    for {
        tick_start := time.Nanoseconds()

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

        sleep_ns := (tick_start + skip_ns) - time.Nanoseconds()
        time.Sleep(sleep_ns)
    }
}

func (g *Game) AddEntity(ent core.Entity) {
    g.ents[ent.Chan()] = ent
    msg := core.MsgEntityAdded{core.NewEntityDesc(ent)}
    g.svc.PubSub <- pubsub.PublishMsg{"entity", msg}
}

func (g *Game) makeEntityList() core.MsgListEntities {
    list := make([]*core.EntityDesc, len(g.ents))

    i := 0
    for _, ent := range g.ents { // Fill the len(g.ents) slots in list
        list[i] = core.NewEntityDesc(ent)
        i++
    }
    return core.MsgListEntities{nil, list}
}

// Returns the next available unique id
func (g *Game) GetUid() core.UniqueId {
    uid := g.nextUid
    g.nextUid++
    return uid
}

// Returns once all services have signalled that they have started.
// Automatically accounts for a variable number of services as contained in the
// ServiceContext struct.
//
// TODO: This implementation is prone to error if the same service sends
// MsgTick more than once, fix?
func (g *Game) waitOnServiceStart(input chan core.Msg) {
    // We can discard the ok value, because svc is always a struct
    val, _ := (reflect.NewValue(g.svc)).(*reflect.StructValue)
    svc_num := val.NumField() - 1 // Don't count Game
    num_left := svc_num

    for {
        msg := <-input
        switch m := msg.(type) {
        case core.MsgTick:
            num_left--
            if num_left == 0 {
                goto done
            }
        default:
            panic("Received message other than MsgTick!")
        }
    }
done:
}

// Starting to get somewhat game specific?

// Creates a new player entity for a requesting client
func (g *Game) spawnPlayer(msg core.MsgSpawnPlayer) {
    if g.PlayerFunc == nil {
        return // Skip spawning a player
    }
    p := g.PlayerFunc(g.GetUid())
    g.AddEntity(p)
    go p.Run(g.svc)
    go func(uid core.UniqueId) {
        desc := core.NewEntityDesc(p)
        msg.Reply <- desc
    }(p.Uid())
}
