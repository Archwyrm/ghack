// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Namespace for component indentification in order to ensure uniqueness
// in bookkeeping.
package cmpId

type StateId int

const (
    State = iota
    EntityList
    Position
)

type ActionId int

const (
    Action = iota
    AddEntity
    Move
)

type EntityId int

const (
    Entity = iota
    Game
    Player
)
