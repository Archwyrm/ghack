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
    "core"
    "game"
    "pubsub"
    "sf/cmpId"
)

// Signal entity's intent to move from point A to point B
type MoveMsg struct {
    Ent *core.EntityDesc // The moving entity
    Vel *s3dm.V3         // The entity's velocity vector
}

// Translates a 3D vector into a cell position. X and Y values are truncated.
func hashV3(vec *s3dm.V3) string {
    return fmt.Sprintf("%d+%d", int(vec.X), int(vec.Y))
}

// Service that controls spatial relations between entities. The world is divided
// into a grid, each of part of the grid is a cell. Currently, only one entity may
// occupy a cell at any given time.
type World struct {
    svc core.ServiceContext
    // Entities may be looked up by position with this
    ents map[string]chan core.Msg
    // Entity position (or cells) as 3D vectors may be looked up with this
    pos map[core.UniqueId]*s3dm.V3
}

func NewWorld(svc core.ServiceContext) *World {
    ents := make(map[string]chan core.Msg)
    pos := make(map[core.UniqueId]*s3dm.V3)
    return &World{svc, ents, pos}
}

func (w *World) Run(input chan core.Msg) {
    // Subscribe to listen for new entities in order to track their position
    w.svc.PubSub <- pubsub.SubscribeMsg{"entity", input}
    w.svc.Game <- core.MsgTick{input} // Service is ready

    for {
        msg := <-input
        switch m := msg.(type) {
        case MoveMsg:
            w.moveEnt(m.Ent, m.Vel)
        case core.MsgEntityAdded:
            reply := make(chan core.State)
            m.Entity.Chan <- core.MsgGetState{cmpId.Position, reply}
            if pos, ok := (<-reply).(Position); ok {
                w.setPos(m.Entity, pos.Position, nil)
            }
        }
    }
}

// Sets the position of an entity within the world. The entity whose position is
// being set is represented by ent. New position is the passed vector new_pos.
// The old position of old_pos is removed if it exists and is not used when nil
// is passed.
func (w *World) setPos(ent *core.EntityDesc, new_pos, old_pos *s3dm.V3) {
    if old_pos != nil {
        w.ents[hashV3(old_pos)] = nil, false // Remove old pos
    }
    w.ents[hashV3(new_pos)] = ent.Chan
    w.pos[ent.Uid] = new_pos
}

func (w *World) moveEnt(ent *core.EntityDesc, vel *s3dm.V3) {
    // Compute new position vector
    old_pos, ok := w.pos[ent.Uid]
    if !ok { // Entity hasn't been added for some reason, bail
        log.Println("No position for", ent.Uid)
        return
    }
    new_pos := old_pos.Add(vel)
    hash := hashV3(new_pos)

    // See if destination cell is occupied
    if _, ok := w.ents[hash]; ok {
        // TODO: HACK remove following if block when spider AI uses time based move
        // Update position if movement is less than 1, this lets spider move slowly
        if math.Fabs(vel.X) < 1 && math.Fabs(vel.Y) < 1 {
            w.pos[ent.Uid] = new_pos
            ent.Chan <- core.MsgSetState{Position{new_pos}}
        }
        return // Can't move there, bail
    }
    // If not, move the entity to the new pos
    w.setPos(ent, new_pos, old_pos)
    // Update entity position state
    ent.Chan <- core.MsgSetState{Position{new_pos}}

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
    reply := make(chan *core.EntityDesc)
    for i := 0; i < count; i++ {
        radius := rand.Float64()*(MAX_DIST-MIN_DIST) + MIN_DIST
        angle := rand.Float64() * 2. * math.Pi
        x := pos.X + radius*math.Cos(angle)
        y := pos.Y + radius*math.Sin(angle)
        // Create the spider entity and position it
        go func(X, Y float64) {
            w.svc.Game <- game.MsgSpawnEntity{InitSpider, reply}
            spider := <-reply
            spider.Chan <- core.MsgSetState{Position{s3dm.NewV3(X, Y, 0.)}}
        }(x, y)
    }
}
