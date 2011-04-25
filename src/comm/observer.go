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
    "reflect"
    "runtime"
    .   "core"
    "pubsub"
    "util"
)

// Signal that an entity should be sent to a client
// TODO: Combine with MsgEntityAdded somehow?
type MsgAddEntity struct {
    Uid  UniqueId
    Name string
}

// Signal that an entity should be removed from a client
// TODO: Combine with MsgEntityRemoved somehow?
type MsgRemoveEntity struct {
    Uid  UniqueId
    Name string
}

// Signal that a state should have its value updated on a client
type MsgUpdateState struct {
    Uid   UniqueId
    State State // Contains Name and Value needed for protocol
}

// Replicates data to a connected client. Views are created for each replicated entity.
// This keeps the game state on the client in sync with the server.
type observer struct {
    svc ServiceContext
    // The client on whose behalf this observer replicates
    client chan Msg
    // Maps the entity to its view's control channel
    // TODO: Store by Uid, not chan?
    views map[chan Msg]chan Msg
    // Channel to control this observer
    ctrl chan Msg
}

// Creates an observer instance in a new goroutine and returns a control channel
func createObserver(svc ServiceContext, client chan Msg) chan Msg {
    // Create struct
    obs := &observer{svc, client, make(map[chan Msg]chan Msg), make(chan Msg)}
    go obs.observe()
    return obs.ctrl
}

// Do initial observer set up
func (obs *observer) init() {
    // Get list of entities for initial sync
    reply := make(chan Msg)
    obs.svc.Game <- MsgListEntities{Reply: reply}
    list, ok := (<-reply).(MsgListEntities)
    if !ok {
        panic("Request received incorrect reply")
    }
    for _, ent := range list.Entities {
        if checkBlacklist(ent.Id) {
            continue
        }
        obs.addView(ent)
    }
    obs.svc.PubSub <- pubsub.SubscribeMsg{"entity", obs.ctrl}
    obs.svc.PubSub <- pubsub.SubscribeMsg{"combat", obs.ctrl}
}

func (obs *observer) observe() {
    obs.init()
    for {
        msg := <-obs.ctrl
        switch m := msg.(type) {
        case MsgTick: // Pass update msg to views
            for _, v := range obs.views {
                v <- msg
            }
        case MsgQuit: // Client has disconnected, shut everything down
            // Views may have pending updates, drain and discard
            go util.DrainUntilQuit(obs.client)
            for _, v := range obs.views {
                v <- msg
            }
            // Close the drain now that all views have gotten quit msg
            obs.client <- msg
            return
        case MsgEntityAdded:
            ent := m.Entity
            if checkBlacklist(ent.Id) {
                continue
            }
            if _, present := obs.views[ent.Chan]; present {
                str := "Duplicate MsgEntityAdded received, entity already has view:\n"
                str = fmt.Sprint(str, ent.Uid, " ", ent.Name)
                panic(str)
            }
            obs.addView(ent)
        case MsgEntityRemoved:
            ent := m.Entity
            // Signal quit to the correct view
            if ch, ok := obs.views[ent.Chan]; ok {
                ch <- MsgQuit{}
            } else {
                str := "Tried to remove an unadded entity:\n"
                str = fmt.Sprint(str, ent.Uid, " ", ent.Name)
                panic(str)
            }
            obs.views[ent.Chan] = nil, false
            obs.client <- MsgRemoveEntity{ent.Uid, ent.Name}
        default:
            obs.eventListener(m)
        }
    }
}

// Creates a new view and starts it replicating
func (obs *observer) addView(ent *EntityDesc) {
    obs.client <- MsgAddEntity{ent.Uid, ent.Name}
    v := &view{client: obs.client, entity: ent.Chan}
    v_ch := make(chan Msg)
    obs.views[ent.Chan] = v_ch
    go v.replicate(ent.Uid, v_ch)
}

// Checks to see if this entity is blacklisted
// Returns true if blacklisted, false otherwise
func checkBlacklist(id EntityId) bool {
    return false // TODO: Actually check
}

// Replicates an individual entity. Each state of the watched entity is tracked
// for changes. If the state has changed then an update is sent to the client.
type view struct {
    client chan Msg
    entity chan Msg
    states StateList // Current value of each replicated state
}

func (v *view) replicate(uid UniqueId, ctrl chan Msg) {
    // TODO: Eventually this list will fill with states that are no longer in
    // the entity, we need a mechanism to clear it out occasionally
    v.states = make(StateList)
    request := MsgGetAllStates{}
    msg := MsgUpdateState{}
    msg.Uid = uid

    for {
        reply := make(chan Msg)
        request.Reply = reply
        select {
        // TODO: White or black list?
        // Get whitelisted states from entity (must check for new states)
        case v.entity <- request:
        case msg := <-ctrl:
            handleCtrl(msg)
        }

        for m := range reply {
            s := m.(State)
            if !checkWhiteList(s.Id()) {
                continue
            }
            // Compare to current value (first time will be none)
            if val, ok := v.states[s.Id()]; ok {
                if reflect.DeepEqual(val, s) {
                    continue
                }
            }
            v.states[s.Id()] = s
            // Send updates for any changed states
            msg.State = s
            v.client <- msg
        }
        // Listen for next update signal
        msg := <-ctrl
        handleCtrl(msg)
    }
}

func handleCtrl(msg Msg) {
    if _, ok := msg.(MsgQuit); ok {
        runtime.Goexit()
    }
}

// Checks to see if this entity is whitelisted
// Returns true if whitelisted, false otherwise
func checkWhiteList(id StateId) bool {
    return true // TODO: Actually check
}

// Send events to client
func (obs *observer) eventListener(msg Msg) {
    obs.client <- msg
}
