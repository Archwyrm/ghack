// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Contains everything necessary for component system

package main

import (
    "cmpId/cmpId"
)

// Holds some kind of state data for a particular named property of an Entity.
// Id() returns a unique ID for each Action (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Action (defined
// by each Entity)
type State interface {
    Id() cmpId.StateId
    Name() string
}

// The Action enacts changes in Entity state. It may be considered a
// transactional state change, or since it arrives by message, a closure.
// Id() returns a unique ID for each Action (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Action (defined
// by each Entity)
type Action interface {
    Id() cmpId.ActionId
    Name() string
    Act(ent Entity)
}

// An Entity is a struct composed from various States and Actions, which each
// make up its data and functionality respectively.
//
// Id() returns a unique ID for each Entity (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Entity (defined
// by each Entity)
// GetState() returns the requested State or nil if it does not exist.
// SetState() sets the value of the passed State within the Entity.
// AddAction() adds the Action to the Entity.
// RemoveAction() removes the Action from the Entity.
type Entity interface {
    Id() cmpId.EntityId
    Name() string
    GetState(state State) State
    SetState(state State)
    AddAction(action Action)
    RemoveAction(action Action)
}

// Simplified declaration
type StateList map[cmpId.StateId]State
// Simplified declaration
type ActionList map[cmpId.ActionId]Action

// Contains all the data that each component needs.
// TODO: Rename to 'Component'?
type CmpData struct {
    // Use maps for easy/add remove for now
    states  StateList
    actions ActionList
    input   chan CmpMsg
}

// Creates a CmpData and initializes its containers.
func NewCmpData() *CmpData {
    states := make(StateList)
    actions := make(ActionList)
    return &CmpData{states, actions, nil}
}

// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Id() cmpId.EntityId { return 0 }
// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Name() string { return "CmpData" }

// The next functions form the core functionality of a component.

// Returns the requested State. TODO: Take StateId?
func (cd CmpData) GetState(state State) State {
    ret := cd.states[state.Id()]
    return ret
}

// Set the value of the passed State. Replaces any existing State that is the same.
func (cd CmpData) SetState(state State) {
    cd.states[state.Id()] = state
}

// Adds to an Entity's actions, causing the Action to be executed on the next tick.
func (cd CmpData) AddAction(action Action) {
    cd.actions[action.Id()] = action
}

// Removes the Action from an Entity's actions.
func (cd CmpData) RemoveAction(action Action) {
    cd.actions[action.Id()] = nil, false
}

// Main loop which handles all component tasks.
func (cd CmpData) Run(input chan CmpMsg) {
    cd.input = input

    for {
        msg := <-input

        // Call the appropriate function based on the msg type
        switch m := msg.(type) {
        case MsgTick:
            cd.update()

        case MsgGetState:
            cd.sendState(m)

        case MsgAddAction:
            cd.AddAction(m.Action)
        }
    }
}

// Loop through each Action and let it run
func (cd CmpData) update() {
    for _, v := range cd.actions {
        v.Act(cd)
    }
}

// Send back the requested State on the provided channel
func (cd CmpData) sendState(msg MsgGetState) {
    state := cd.states[msg.StateId]
    msg.StateReply <- state
}
