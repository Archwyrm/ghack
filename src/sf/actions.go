// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    "github.com/tm1rbrt/s3dm"
    "core"
    "cmpId"
)

type Move struct {
    Direction *s3dm.V3
}

func (a Move) Id() core.ActionId { return cmpId.Move }
func (a Move) Name() string      { return "Move" }

// Modifies the Position of an Entity with the passed Move vector.
func (a Move) Act(ent core.Entity) {
    // Automatically create a position if it does not exist, keep?
    pos, ok := ent.GetState(cmpId.Position).(Position)
    if !ok {
        pos = Position{&s3dm.V3{0, 0, 0}}
        ent.SetState(pos)
    }
    pos.Position = pos.Position.Add(a.Direction)

    // Stick the value back in
    // This possibly orphans the existing structure which may be hard on the GC
    // if we are doing this for many entities every frame, pointers may be helpful..
    // Still, the actual state setting should be a message, not direct access so we
    // can run Actions in goroutines
    ent.SetState(pos)
}
