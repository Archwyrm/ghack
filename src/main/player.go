package main

import (
    "cmpId/cmpId"
)

type Player struct {
    *CmpData
}

func (p Player) Id() int      { return cmpId.Player }
func (p Player) Name() string { return "Player" }

func NewPlayer() *Player {
    return &Player{NewCmpData()}
}
