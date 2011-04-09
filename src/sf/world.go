// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    "fmt"
    "log"
    "github.com/tm1rbrt/s3dm"
    "core"
    "pubsub"
    "sf/cmpId"
)

// Signal entity's intent to move from point A to point B
type MoveMsg struct {
    Ent *core.EntityDesc // The moving entity
    Vel *s3dm.V3         // The entity's velocity vector
}

// The world is divided into a grid, each of part of the grid is a cell
type cell struct {
    x, y int
}

// Returns true if both cells have the same position, false if otherwise
func (c cell) Equals(rhs cell) bool {
    return c.x == rhs.x && c.y == rhs.y
}

func (c cell) Hash() string {
    return fmt.Sprintf("%d+%d", c.x, c.y)
}

func newCell(vec *s3dm.V3) *cell {
    return &cell{int(vec.X), int(vec.Y)}
}

type World struct {
    svc   core.ServiceContext
    // Entities may be looked up by position with this map
    ents  map[string]chan core.Msg
    // Cells (or entity position) may be looked up by entity with this map
    cells map[core.UniqueId]*cell
}

func NewWorld(svc core.ServiceContext) *World {
    ents := make(map[string]chan core.Msg)
    cells := make(map[core.UniqueId]*cell)
    return &World{svc, ents, cells}
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
                cell := newCell(pos.Position)
                w.ents[cell.Hash()] = m.Entity.Chan
                w.cells[m.Entity.Uid] = cell
            }
        }
    }
}

func (w *World) moveEnt(ent *core.EntityDesc, vel *s3dm.V3) {
    // Compute new position vector
    old_cell, ok := w.cells[ent.Uid]
    cell := old_cell
    if !ok { // Entity hasn't been added for some reason, bail
        log.Println("No position for", ent.Uid)
        return
    }
    vel_cell := newCell(vel) // Convert to cell to truncate vector values
    cell.x += vel_cell.x
    cell.y += vel_cell.y

    hash := cell.Hash() // TODO: Dump cell and just hash vec3?

    // See if this cell is occupied
    if _, ok := w.ents[hash]; ok {
        return // Can't move there, bail
    }
    // If not, move the entity to the new cell
    w.ents[old_cell.Hash()] = nil, false // Remove old cell
    w.ents[hash] = ent.Chan
    w.cells[ent.Uid] = cell
    // Update entity position state
    new_v3 := s3dm.NewV3(float64(cell.x), float64(cell.y), 0)
    ent.Chan <- core.MsgSetState{Position{new_v3}}
}
