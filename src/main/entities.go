// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "core"
    "cmpId"
)

// A plain component definition needs only four (reasonably) compact lines
type Player struct {
    *core.CmpData
}

func (p Player) Id() core.EntityId        { return cmpId.Player }
func (p Player) Name() string             { return "Player" }
func NewPlayer(uid core.UniqueId) *Player { return &Player{core.NewCmpData(uid)} }

type Spider struct {
    *core.CmpData
}

func (p Spider) Id() core.EntityId        { return cmpId.Spider }
func (p Spider) Name() string             { return "Spider" }
func NewSpider(uid core.UniqueId) *Spider { return &Spider{core.NewCmpData(uid)} }
