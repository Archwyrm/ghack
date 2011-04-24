// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Spider Forest package. Contains all code specific to the Spider Forest gameplay.
package sf

import (
    "github.com/tm1rbrt/s3dm"
    "core"
    "sf/cmpId"
)

type Player struct {
    *core.CmpData
}

func InitPlayer(uid core.UniqueId) core.Entity {
    p := &Player{core.NewCmpData(uid, cmpId.Player, "Player")}
    p.SetState(Position{&s3dm.V3{1, 1, 0}})
    p.SetState(Asset{"@"})
    p.SetState(Health{10})
    p.SetState(MaxHealth{10})
    return p
}

type Spider struct {
    *core.CmpData
}

func InitSpider(uid core.UniqueId) core.Entity {
    s := &Spider{core.NewCmpData(uid, cmpId.Spider, "Spider")}
    s.SetState(Position{&s3dm.V3{1, 1, 0}})
    s.SetState(Asset{"s"})
    s.SetState(Health{4})
    s.SetState(MaxHealth{4})
    return s
}
