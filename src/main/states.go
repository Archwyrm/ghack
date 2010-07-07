package main

import (
    "cmpId/cmpId"
)

// Contains a list of entities. Note that these are entire entities,
// not shared references. Currently used by the Game entity to contain all
// other entities.
type EntityList struct {
    Entities map[string]Entity
}

func (p EntityList) Id() int      { return cmpId.EntityList }
func (p EntityList) Name() string { return "EntityList" }

func NewEntityList() *EntityList {
    return &EntityList{make(map[string]Entity)}
}

// Simple 2D position
type Position struct {
    X, Y int
}

func (p Position) Id() int      { return cmpId.Position }
func (p Position) Name() string { return "Position" }
