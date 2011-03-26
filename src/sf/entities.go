// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Spider Forest package. Contains all code specific to the Spider Forest gameplay.
package sf

import (
    "github.com/tm1rbrt/s3dm"
    "core"
    "cmpId"
)

type Player struct {
    *core.CmpData
}

func (p Player) Id() core.EntityId { return cmpId.Player }
func (p Player) Name() string      { return "Player" }

func NewPlayer(uid core.UniqueId) *Player {
    p := &Player{core.NewCmpData(uid)}
    p.SetState(Position{&s3dm.V3{1, 1, 0}})
    p.SetState(Asset{"@"})
    p.SetState(Health{10})
    p.SetState(MaxHealth{10})
    p.SetState(KillCount{0})
    return p
}

// A plain component definition needs only four (reasonably) compact lines
type Spider struct {
    *core.CmpData
}

func (p Spider) Id() core.EntityId { return cmpId.Spider }
func (p Spider) Name() string      { return "Spider" }

func NewSpider(uid core.UniqueId) *Spider {
    s := &Spider{core.NewCmpData(uid)}
    s.SetState(Position{&s3dm.V3{1, 1, 0}})
    s.SetState(Asset{"s"})
    s.SetState(Health{10})
    s.SetState(MaxHealth{10})
    return s
}