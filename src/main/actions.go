// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "core/core"
    "cmpId/cmpId"
)

// Adds the passed Entity to an Entity's EntityList. Used by Game to
// populate its list.
type AddEntity struct {
    newEntity core.Entity
}

func (a AddEntity) Id() core.ActionId               { return cmpId.AddEntity }
func (a AddEntity) Name() string                    { return "AddEntity" }
func NewAddEntity(newEntity core.Entity) *AddEntity { return &AddEntity{newEntity} }

func (a AddEntity) Act(ent core.Entity) {
    var list EntityList
    list = ent.GetState(list).(EntityList)
    list.Entities[a.newEntity.Name()] = a.newEntity
}

// Simple movement.
type Move struct {
    DirX int
    DirY int
}

func (a Move) Id() core.ActionId { return cmpId.Move }
func (a Move) Name() string      { return "Move" }

// Modifies the Position of an Entity with the passed Move vector.
func (a Move) Act(ent core.Entity) {
    var pos Position
    var ok bool
    // Automatically create a position if it does not exist, keep?
    pos, ok = ent.GetState(pos).(Position)
    if !ok {
        pos := Position{0, 0}
        ent.SetState(pos)
    }
    pos.X += a.DirX
    pos.Y += a.DirY

    // Stick the value back in
    // This possibly orphans the existing structure which may be hard on the GC
    // if we are doing this for many entities every frame, pointers may be helpful..
    // Still, the actual state setting should be a message, not direct access so we
    // can run Actions in goroutines
    ent.SetState(pos)
}
