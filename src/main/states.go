// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "core/core"
    "cmpId/cmpId"
    "github.com/tm1rbrt/s3dm"
)

// Contains a list of entities. Note that these are entire entities,
// not shared references. Currently used by the Game entity to contain all
// other entities.
type EntityList struct {
    Entities map[chan core.Msg]core.Entity
}

func (p EntityList) Id() core.StateId { return cmpId.EntityList }
func (p EntityList) Name() string     { return "EntityList" }

func NewEntityList() EntityList {
    return EntityList{make(map[chan core.Msg]core.Entity)}
}

type Position struct {
    Position *s3dm.V3
}

func (p Position) Id() core.StateId { return cmpId.Position }
func (p Position) Name() string     { return "Position" }
