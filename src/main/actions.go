package main

import (
    "cmpId/cmpId"
)

// Adds the passed Entity to an Entity's EntityList. Used by Game to
// populate its list.
type AddEntity struct {
    newEntity Entity
}

func (a AddEntity) Id() cmpId.ActionId         { return cmpId.AddEntity }
func (a AddEntity) Name() string               { return "AddEntity" }
func NewAddEntity(newEntity Entity) *AddEntity { return &AddEntity{newEntity} }

func (a AddEntity) Act(ent Entity) {
    var list EntityList
    list = ent.GetState(list).(EntityList)
    list.Entities[a.newEntity.Name()] = a.newEntity
}

// Simple movement.
type Move struct {
    DirX int
    DirY int
}

func (a Move) Id() cmpId.ActionId { return cmpId.Move }
func (a Move) Name() string       { return "Move" }

// Modifies the Position of an Entity with the passed Move vector.
func (a Move) Act(ent Entity) {
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
