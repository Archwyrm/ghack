// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    "fmt"
    "log"
    "math"
    "rand"
    "github.com/tm1rbrt/s3dm"
    .   "core"
    "game"
    "pubsub"
    "sf/cmpId"
)

// Signal entity's intent to move from point A to point B
type MoveMsg struct {
    Ent *EntityDesc // The moving entity
    Vel *s3dm.V3    // The entity's velocity vector
}

// Translates a 3D vector into a cell position. X and Y values are truncated.
func hashV3(vec *s3dm.V3) string {
    return fmt.Sprintf("%d+%d", int(vec.X), int(vec.Y))
}

// Service that controls spatial relations between entities. The world is divided
// into a grid, each of part of the grid is a cell. Currently, only one entity may
// occupy a cell at any given time.
type World struct {
    *HandlerQueue
    svc ServiceContext
    // Entities may be looked up by position with this
    ents map[string]chan Msg
    // Entity position (or cells) as 3D vectors may be looked up with this
    pos map[UniqueId]*s3dm.V3
    // Listens on this channel to receive messages
    input chan Msg
}

func NewWorld(svc ServiceContext) *World {
    hq := NewHandlerQueue()
    ents := make(map[string]chan Msg)
    pos := make(map[UniqueId]*s3dm.V3)
    return &World{hq, svc, ents, pos, nil}
}

func (w *World) Chan() chan Msg { return w.input }

func (w *World) Run(input chan Msg) {
    // Subscribe to listen for new entities in order to track their position
    Send(w, w.svc.PubSub, pubsub.SubscribeMsg{"entity", input})
    Send(w, w.svc.Game, MsgTick{input}) // Service is ready

    for {
        w.handle(w.GetMsg(input))
    }
}

func (w *World) handle(msg Msg) {
    switch m := msg.(type) {
    case MoveMsg:
        w.moveEnt(m.Ent, m.Vel)
    case MsgEntityAdded:
        reply := make(chan Msg)
        Send(w, m.Entity.Chan, MsgGetState{cmpId.Position, reply})
        if pos, ok := Recv(w, reply).(Position); ok {
            w.putInEmptyPos(m.Entity, pos.Position)
        }
    case MsgEntityRemoved:
        pos := w.pos[m.Entity.Uid]
        w.pos[m.Entity.Uid] = nil, false
        w.ents[hashV3(pos)] = nil, false
    }
}

// Puts the passed entity in an empty position as close to pos as possible.
// TODO: Current implementation doesn't try very hard at closeness ;)
func (w *World) putInEmptyPos(ent *EntityDesc, pos *s3dm.V3) {
    old_pos := pos.Copy()
    for {
        if _, ok := w.ents[hashV3(pos)]; !ok {
            break // Empty, proceed
        }
        inc := &s3dm.V3{1, 1, 0}
        pos = pos.Add(inc)
    }
    w.setPos(ent, pos, nil)
    if !pos.Equals(old_pos) {
        // Update with new pos
        Send(w, ent.Chan, MsgSetState{Position{pos}})
    }
}

// Sets the position of an entity within the world. The entity whose position is
// being set is represented by ent. New position is the passed vector new_pos.
// The old position of old_pos is removed if it exists and is not used when nil
// is passed.
func (w *World) setPos(ent *EntityDesc, new_pos, old_pos *s3dm.V3) {
    if old_pos != nil {
        w.ents[hashV3(old_pos)] = nil, false // Remove old pos
    }
    w.ents[hashV3(new_pos)] = ent.Chan
    w.pos[ent.Uid] = new_pos
}

func (w *World) moveEnt(ent *EntityDesc, vel *s3dm.V3) {
    // Compute new position vector
    old_pos, ok := w.pos[ent.Uid]
    if !ok { // Entity hasn't been added for some reason, bail
        log.Println("No position for", ent.Uid)
        return
    }
    new_pos := old_pos.Add(vel)
    hash := hashV3(new_pos)

    // See if destination cell is occupied
    if ent_ch, ok := w.ents[hash]; ok {
        // TODO: HACK remove following if block when spider AI uses time based move
        // Update position if movement is less than 1, this lets spider move slowly
        if math.Fabs(vel.X) < 1 && math.Fabs(vel.Y) < 1 {
            w.pos[ent.Uid] = new_pos
            Send(w, ent.Chan, MsgSetState{Position{new_pos}})
        }
        // Can't move there, attack instead
        Send(w, ent_ch, MsgRunAction{Attack{ent}, false})
        return
    }
    // If not, move the entity to the new pos
    w.setPos(ent, new_pos, old_pos)
    // Update entity position state
    Send(w, ent.Chan, MsgSetState{Position{new_pos}})

    // Spawn spiders as players move around
    if ent.Id == cmpId.Player {
        w.spawnSpiders(new_pos)
    }
}

func (w *World) spawnSpiders(pos *s3dm.V3) {
    // Represents the gradually increasing difficulty as the player gets
    // farther from the center of Spider Forest. Every time the player
    // passes this threshold, another spider will definitely spawn upon
    // player movement. So if the player is 173 units from the center,
    // 173/50=3 spiders will spawn with a 23/50 chance to make it 4.
    // The lower this number, the more difficult the game will be.
    const LEVEL_DIST = 100.
    // Minimum and maximum distance from the player to spawn spiders
    const MIN_DIST = 40.
    const MAX_DIST = 60.

    // Find how many spiders will definitely spawn
    dist := pos.Length()
    count := 0
    for dist >= LEVEL_DIST {
        dist -= LEVEL_DIST
        count++
    }
    // Random chance for a spider to spawn
    if rand.Float64() <= dist/LEVEL_DIST {
        count++
    }

    // Spawn spiders at a random point in a ring around the player
    // between MIN_DIST and MAX_DIST.
    reply := make(chan Msg)
    for i := 0; i < count; i++ {
        radius := rand.Float64()*(MAX_DIST-MIN_DIST) + MIN_DIST
        angle := rand.Float64() * 2. * math.Pi
        x := pos.X + radius*math.Cos(angle)
        y := pos.Y + radius*math.Sin(angle)
        // Create the spider entity and position it
        Send(w, w.svc.Game, game.MsgSpawnEntity{InitSpider, reply})
        spider := Recv(w, reply).(*EntityDesc)
        Send(w, spider.Chan, MsgSetState{Position{s3dm.NewV3(x, y, 0.)}})
    }
}
