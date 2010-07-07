package main

import (
    "cmpId/cmpId"
)

type EntityList struct {
    Entities map[string]Entity
}

func (p EntityList) Id() int      { return cmpId.EntityList }
func (p EntityList) Name() string { return "EntityList" }

func NewEntityList() *EntityList {
    return &EntityList{make(map[string]Entity)}
}

type Position struct {
    X, Y int
}

func (p Position) Id() int      { return cmpId.Position }
func (p Position) Name() string { return "Position" }
