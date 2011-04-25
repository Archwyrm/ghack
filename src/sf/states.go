// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    .   "core"
    "sf/cmpId"
    "github.com/tm1rbrt/s3dm"
)

type Position struct {
    Position *s3dm.V3
}

func (p Position) Id() StateId  { return cmpId.Position }
func (p Position) Name() string { return "Position" }

// The asset used to represent (or draw) the entity to which this state belongs.
type Asset struct {
    Asset string
}

func (x Asset) Id() StateId  { return cmpId.Asset }
func (x Asset) Name() string { return "Asset" }

type Health struct {
    Health float32
}

func (x Health) Id() StateId  { return cmpId.Health }
func (x Health) Name() string { return "Health" }

type MaxHealth struct {
    MaxHealth float32
}

func (x MaxHealth) Id() StateId  { return cmpId.MaxHealth }
func (x MaxHealth) Name() string { return "MaxHealth" }
