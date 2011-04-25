// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    "github.com/tm1rbrt/s3dm"
    .   "core"
    "sf/cmpId"
    "util"
    "pubsub"
)

type Move struct {
    Direction *s3dm.V3
}

func (a Move) Id() ActionId { return cmpId.Move }
func (a Move) Name() string { return "Move" }

// Modifies the Position of an Entity with the passed Move vector.
func (a Move) Act(ent Entity, svc ServiceContext) {
    svc.World <- MoveMsg{NewEntityDesc(ent), a.Direction}
}

// Does damage to the calling entity, the entity being attacked.
// Removes the entity if Health is zero.
type Attack struct {
    Attacker *EntityDesc
}

func (a Attack) Id() ActionId { return cmpId.Attack }
func (a Attack) Name() string { return "Attack" }

func (a Attack) Act(ent Entity, svc ServiceContext) {
    var health Health
    var ok bool
    if health, ok = (ent.GetState(cmpId.Health)).(Health); !ok {
        return // Ent has not Health state
    }
    health.Health-- // Extremely complex damage formula
    ed := NewEntityDesc(ent)
    util.Send(svc.PubSub, pubsub.PublishMsg{"combat", MsgCombatHit{a.Attacker, ed, 1}})
    if health.Health <= 0 {
        ent.SetState(Remove{})
        util.Send(svc.Game, MsgEntityRemoved{ed})
        util.Send(svc.PubSub, pubsub.PublishMsg{"combat", MsgEntityDeath{ed, a.Attacker}})
    }
    ent.SetState(health)
}
