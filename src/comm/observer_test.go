// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "testing"
    "time"
    "core/core"
)

/*var map[string]core.Entity = {
    "Hypothetical Entity 1" : Entity?
}*/

type testEntity struct {
    *core.CmpData
}

func (x testEntity) Id() core.EntityId { return 0 }
func (x testEntity) Name() string      { return "TestEntity" }
func NewTestEntity() *testEntity       { return &testEntity{core.NewCmpData()} }

type testState struct {
    Value int
}

func (x testState) Id() core.StateId { return 0 }
func (x testState) Name() string     { return "TestState" }

// Test replicating entity data through observers up through the initial sync.
func TestObserver(t *testing.T) {
    svc := core.NewServiceContext()
    ent, ent_ch := createTestEntity(1)
    go gameEmulator(t, svc, ent_ch, ent.Id(), ent.Name())
    go pubsubEmulator(t, svc)
    client := make(chan core.Msg)
    obs := createObserver(svc, client)

    // Expecting one entity added
    msg := getMessage(t, client)
    if _, ok := msg.(MsgAddEntity); !ok {
        t.Fatal("No entity added")
    }

    // Expecting one state update
    msg = getMessage(t, client)
    if m, ok := msg.(MsgUpdateState); !ok {
        t.Fatal("No state update received")
    } else {
        // TODO: Check Id
        var state testState
        state = ent.GetState(state.Id()).(testState)
        mstate, ok := m.State.(testState)
        if !ok {
            t.Fatalf("State update did not contain a testState")
        }
        if state.Value != mstate.Value {
            t.Fatalf("testState values do not match")
        }
    }

    // TODO: Send quit message to obs
    obs <- false // Quit
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
    ent := NewTestEntity()
    ent.SetState(testState{value})
    ch := make(chan core.Msg)
    go ent.Run(ch)
    return ent, ch
}

// Masquerades as a game entity for testing purposes
func gameEmulator(t *testing.T, svc core.ServiceContext, testEnt chan core.Msg,
testId core.EntityId, testName string) {
    list, ok := (<-svc.Game).(core.MsgListEntities)
    if !ok {
        t.Fatal("Unexpected message sent to game service")
    }

    list.Reply <- core.MsgListEntities{nil,
        []chan core.Msg{testEnt},
        []core.EntityId{testId},
        []string{testName}}
}

// Masquerades as a pubsub service for testing purposes
func pubsubEmulator(t *testing.T, svc core.ServiceContext) {
    // TODO: Handle data. Just drain the channel for now.
    <-svc.PubSub
    t.Fatal("Not implemented yet!")
}
