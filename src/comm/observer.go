// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

// The idea of using observers and views is to keep a strict separation between
// networking and game code. This portion of the networking is aware of the game
// code, but does not treat it in any specific way. However, on the game code
// side, Entities are unaware that they are not talking to other entities.

import (
    "fmt"
    "core/core"
    "pubsub/pubsub"
)

// Signal that an entity should be sent to a client
// TODO: Combine with core.MsgEntityAdded somehow?
type MsgAddEntity struct {
    Id   int32 // Unique Id
    Name string
}

// Signal that an entity should be removed from a client
// TODO: Combine with core.MsgEntityRemoved somehow?
type MsgRemoveEntity struct {
    Id   int32 // Unique Id
    Name string
}

// Signal that a state should have its value updated on a client
type MsgUpdateState struct {
    Id    int32      // Unique Id
    State core.State // Contains Name and Value needed for protocol
}

// Replicates data to a connected client. Views are created for each replicated entity.
// This keeps the game state on the client in sync with the server.
type observer struct {
    svc core.ServiceContext
    // The client on whose behalf this observer replicates
    client chan core.Msg
    // Maps the entity to its view's control channel
    // TODO: Store by Id, not chan?
    views map[chan core.Msg]chan core.Msg
    // Channel to control this observer
    ctrl chan core.Msg
}

// Creates an observer instance in a new goroutine and returns a control channel
func createObserver(svc core.ServiceContext, client chan core.Msg) chan core.Msg {
    // Create struct
    obs := &observer{svc, client, make(map[chan core.Msg]chan core.Msg), make(chan core.Msg)}
    go obs.observe()
    return obs.ctrl
}

// Do initial observer set up
func (obs *observer) init() {
    // Get list of entities for initial sync
    reply := make(chan core.Msg)
    obs.svc.Game <- core.MsgListEntities{Reply: reply}
    list, ok := (<-reply).(core.MsgListEntities)
    if !ok {
        panic("Request received incorrect reply")
    }
    for i := range list.Entities {
        if checkBlacklist(list.Types[i]) {
            continue
        }
        // TODO: set proper id
        obs.addView(0, list.Entities[i], list.Names[i])
    }
    obs.svc.PubSub <- pubsub.SubscribeMsg{"entity", obs.ctrl}
}

func (obs *observer) observe() {
    obs.init()
    // TODO: Listen for entities added or removed
    for {
        msg := <-obs.ctrl
        switch m := msg.(type) {
        case core.MsgTick: // Pass update msg to views
            for _, v := range obs.views {
                v <- msg
            }
        case core.MsgEntityAdded:
            if checkBlacklist(0) { // TODO: Check proper type id
                continue
            }
            if _, present := obs.views[m.Entity]; present {
                str := "Duplicate MsgEntityAdded received, entity already has view:\n"
                str = fmt.Sprint(str, m.Id, " ", m.Name)
                panic(str)
            }
            obs.addView(m.Id, m.Entity, m.Name)
        case core.MsgEntityRemoved:
            // Signal quit to the correct view
            if ch, ok := obs.views[m.Entity]; ok {
                ch <- core.MsgQuit{}
            } else {
                str := "Tried to remove an unadded entity:\n"
                str = fmt.Sprint(str, m.Id, " ", m.Name)
                panic(str)
            }
            obs.views[m.Entity] = nil, false
            obs.client <- MsgRemoveEntity{int32(m.Id), m.Name}
        }
    }
}

// Creates a new view and starts it replicating
func (obs *observer) addView(id int, entity chan core.Msg, entName string) {
    obs.client <- MsgAddEntity{int32(id), entName}
    v := &view{client: obs.client, entity: entity}
    v_ch := make(chan core.Msg)
    obs.views[entity] = v_ch
    go v.replicate(v_ch)
}

// Checks to see if this entity is blacklisted
// Returns true if blacklisted, false otherwise
func checkBlacklist(id core.EntityId) bool {
    return false // TODO: Actually check
}

// Replicates an individual entity. Each state of the watched entity is tracked
// for changes. If the state has changed then an update is sent to the client.
type view struct {
    client chan core.Msg
    entity chan core.Msg
    states core.StateList // Current value of each replicated state
}

func (v *view) replicate(ctrl chan core.Msg) {
    v.states = make(core.StateList)
    request := core.MsgGetAllStates{}
    // TODO: Set Id as this is always the same for a given view
    msg := MsgUpdateState{}

    for {
        reply := make(chan core.State)
        request.StateReply = reply
        // TODO: White or black list?
        // Get whitelisted states from entity (must check for new states)
        v.entity <- request
        for s := range reply {
            if !checkWhiteList(s.Id()) {
                continue
            }
            // Compare to current value (first time will be none)
            if v, ok := v.states[s.Id()]; ok {
                // TODO: Use reflection to compare values of states
                // For now, always update
                _ = v
            }
            // Send updates for any changed states
            msg.State = s
            v.client <- msg
        }
        // Listen for next update signal
        msg := <-ctrl
        if _, ok := msg.(core.MsgQuit); ok {
            return
        }
    }
}

// Checks to see if this entity is whitelisted
// Returns true if whitelisted, false otherwise
func checkWhiteList(id core.StateId) bool {
    return true // TODO: Actually check
}
