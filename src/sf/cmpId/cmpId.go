// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Namespace for component indentification in order to ensure uniqueness
// in bookkeeping.
package cmpId

// States
const (
    State = iota
    EntityList
    Position
    Asset
    Health
    MaxHealth
    KillCount
)

// Actions
const (
    Action = iota
    AddEntity
    Move
)

// Entities
const (
    Entity = iota
    Player
    Spider
)
