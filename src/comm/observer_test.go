// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "testing"
    "time"
    "core"
    "pubsub"
)

type testEntity struct {
    *core.CmpData
}

func (x testEntity) Id() core.EntityId { return 0 }
func (x testEntity) Name() string      { return "TestEntity" }

func NewTestEntity(uid core.UniqueId) *testEntity {
    return &testEntity{core.NewCmpData(uid)}
}

type testState struct {
    Value int
}

func (x testState) Id() core.StateId { return 0 }
func (x testState) Name() string     { return "TestState" }

var nextUid core.UniqueId = 1

// Test replicating entity data through observers up through the initial sync.
func TestObserver(t *testing.T) {
    svc := core.NewServiceContext()
    ent, ent_ch := createTestEntity(1)
    go gameEmulator(t, svc, ent_ch, ent)
    go pubsubEmulator(t, svc)
    client := make(chan core.Msg)
    obs := createObserver(svc, client)

    // Expecting one entity added
    verifyEntityAdded(t, client, ent)
    // Expecting one state update
    verifyStateUpdated(t, client, ent)

    // Create entity and publish its addition
    ent2, ent_ch2 := createTestEntity(2)
    desc := core.NewEntityDesc(ent2, ent_ch2)
    add_msg := core.MsgEntityAdded{desc}
    svc.PubSub <- pubsub.PublishMsg{"entity", add_msg}

    // Expecting entity added and one state updated
    verifyEntityAdded(t, client, ent2)
    verifyStateUpdated(t, client, ent2)

    // Expecting entity removed
    rm_msg := core.MsgEntityRemoved{desc}
    svc.PubSub <- pubsub.PublishMsg{"entity", rm_msg}
    verifyEntityRemoved(t, client, ent2)

    obs <- core.MsgQuit{}
}

func TestDuplicateEntity(t *testing.T) {
    // TODO: Implement trying to add same entity twice (observer should panic)
}

func TestRemovingUnaddedEntity(t *testing.T) {
    // TODO: Implement trying to remove an entity that has not been added
    // (observer should panic)
}

func verifyEntityAdded(t *testing.T, client chan core.Msg, ent core.Entity) {
    msg := getMessage(t, client)
    if m, ok := msg.(MsgAddEntity); !ok {
        t.Fatal("No entity added")
    } else {
        if ent.Uid() != m.Uid {
            t.Errorf("Entity added: Uids do not match: Ent %v != Msg %v", ent.Uid(), m.Uid)
        }
        if ent.Name() != m.Name {
            t.Errorf("Entity added: Types do not match: Ent %s != Msg %s", ent.Name(), m.Name)
        }
    }
}

func verifyEntityRemoved(t *testing.T, client chan core.Msg, ent core.Entity) {
    msg := getMessage(t, client)
    if m, ok := msg.(MsgRemoveEntity); !ok {
        t.Fatal("No entity removed")
    } else {
        if ent.Uid() != m.Uid {
            t.Errorf("Entity removed: Uids do not match: Ent %v != Msg %v", ent.Uid(), m.Uid)
        }
        if ent.Name() != m.Name {
            t.Errorf("Entity removed: Types do not match: Ent %s != Msg %s", ent.Name(), m.Name)
        }
    }
}

func verifyStateUpdated(t *testing.T, client chan core.Msg, ent core.Entity) {
    msg := getMessage(t, client)
    if m, ok := msg.(MsgUpdateState); !ok {
        t.Fatal("No state update received")
    } else {
        if ent.Uid() != m.Uid {
            t.Errorf("State update: Uids do not match: Ent %v != Msg %v", ent.Uid(), m.Uid)
        }
        var state testState
        state = ent.GetState(state.Id()).(testState)
        mstate, ok := m.State.(testState)
        if !ok {
            t.Fatalf("State update did not contain a testState")
        }
        if state.Value != mstate.Value {
            t.Fatalf("testState values do not match, state: %v\n message state:%v",
                state.Value, mstate.Value)
        }
    }
}

// Gets a message or times out with an error
func getMessage(t *testing.T, ch chan core.Msg) core.Msg {
    select {
    case msg := <-ch:
        return msg
    case <-time.After(1e8): // After 100 ms, timeout
        t.Fatalf("No replication data sent!")
    }
    return nil // Should never be reached
}

func createTestEntity(value int) (core.Entity, chan core.Msg) {
    ent := NewTestEntity(nextUid)
    nextUid++
    ent.SetState(testState{value})
    ch := make(chan core.Msg)
    go ent.Run(ch)
    return ent, ch
}

// Masquerades as a game entity for testing purposes
func gameEmulator(t *testing.T, svc core.ServiceContext, testCh chan core.Msg,
testEnt core.Entity) {
    list, ok := (<-svc.Game).(core.MsgListEntities)
    if !ok {
        t.Fatal("Unexpected message sent to game service")
    }

    desc := []*core.EntityDesc{core.NewEntityDesc(testEnt, testCh)}
    list.Reply <- core.MsgListEntities{nil, desc}
}

// Masquerades as a pubsub service for testing purposes
func pubsubEmulator(t *testing.T, svc core.ServiceContext) {
    var obs chan core.Msg
    for {
        msg := <-svc.PubSub
        switch m := msg.(type) {
        case pubsub.SubscribeMsg:
            if m.Topic != "entity" {
                t.Fatalf("Observer subscribed to wrong topic: %s", m.Topic)
            }
            obs = m.ReplyChan
        case pubsub.PublishMsg:
            if obs == nil {
                t.Fatal("Observer not subscribed!")
            }
            obs <- m.Data
        default:
            t.Fatal("Something wrong with test")
        }
    }
}
