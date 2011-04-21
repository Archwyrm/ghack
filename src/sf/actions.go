// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    "github.com/tm1rbrt/s3dm"
    "core"
    "sf/cmpId"
)

type Move struct {
    Direction *s3dm.V3
}

func (a Move) Id() core.ActionId { return cmpId.Move }
func (a Move) Name() string      { return "Move" }

// Modifies the Position of an Entity with the passed Move vector.
func (a Move) Act(ent core.Entity, svc core.ServiceContext) {
    svc.World <- MoveMsg{core.NewEntityDesc(ent), a.Direction}
}
