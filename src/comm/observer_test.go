// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "testing"
    "time"
    .   "core"
    "pubsub"
)

func InitTestEntity(uid UniqueId) Entity {
    return NewCmpData(uid, 0, "TestEntity")
}

type testState struct {
    Value int
}

func (x testState) Id() StateId  { return 0 }
func (x testState) Name() string { return "TestState" }

var nextUid UniqueId = 1

// Test replicating entity data through observers up through the initial sync.
func TestObserver(t *testing.T) {
    svc := NewServiceContext()
    ent := createTestEntity(svc, 1)
    go gameEmulator(t, svc, ent.Chan(), ent)
    go pubsubEmulator(t, svc)
    client := make(chan Msg)
    obs := createObserver(svc, client)

    // Expecting one entity added
    verifyEntityAdded(t, client, ent)
    // Expecting one state update
    verifyStateUpdated(t, client, ent)

    // Create entity and publish its addition
    ent2 := createTestEntity(svc, 2)
    desc := NewEntityDesc(ent2)
    add_msg := MsgEntityAdded{desc}
    svc.PubSub <- pubsub.PublishMsg{"entity", add_msg}

    // Expecting entity added and one state updated
    verifyEntityAdded(t, client, ent2)
    verifyStateUpdated(t, client, ent2)

    // Expecting entity removed
    rm_msg := MsgEntityRemoved{desc}
    svc.PubSub <- pubsub.PublishMsg{"entity", rm_msg}
    verifyEntityRemoved(t, client, ent2)

    obs <- MsgQuit{}
}

func TestDuplicateEntity(t *testing.T) {
    // TODO: Implement trying to add same entity twice (observer should panic)
}

func TestRemovingUnaddedEntity(t *testing.T) {
    // TODO: Implement trying to remove an entity that has not been added
    // (observer should panic)
}

func verifyEntityAdded(t *testing.T, client chan Msg, ent Entity) {
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

func verifyEntityRemoved(t *testing.T, client chan Msg, ent Entity) {
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

func verifyStateUpdated(t *testing.T, client chan Msg, ent Entity) {
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
func getMessage(t *testing.T, ch chan Msg) Msg {
    select {
    case msg := <-ch:
        return msg
    case <-time.After(1e8): // After 100 ms, timeout
        t.Fatalf("No replication data sent!")
    }
    return nil // Should never be reached
}

func createTestEntity(svc ServiceContext, value int) Entity {
    ent := InitTestEntity(nextUid)
    nextUid++
    ent.SetState(testState{value})
    go ent.Run(svc)
    return ent
}

// Masquerades as a game entity for testing purposes
func gameEmulator(t *testing.T, svc ServiceContext, testCh chan Msg,
testEnt Entity) {
    list, ok := (<-svc.Game).(MsgListEntities)
    if !ok {
        t.Fatal("Unexpected message sent to game service")
    }

    desc := []*EntityDesc{NewEntityDesc(testEnt)}
    list.Reply <- MsgListEntities{nil, desc}
}

// Masquerades as a pubsub service for testing purposes
func pubsubEmulator(t *testing.T, svc ServiceContext) {
    var obs chan Msg
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
