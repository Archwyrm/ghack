// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Contains everything necessary for component system

package core

// Identifies a single state type
type StateId int
// Identifies a single action type
type ActionId int
// Identifies a single entity type
type EntityId int

// Holds some kind of state data for a particular named property of an Entity.
// Id() returns a unique ID for each Action (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Action (defined
// by each Entity)
type State interface {
    Id() StateId
    Name() string
}

// The Action enacts changes in Entity state. It may be considered a
// transactional state change, or since it arrives by message, a closure.
// Id() returns a unique ID for each Action (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Action (defined
// by each Entity)
type Action interface {
    Id() ActionId
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
    Id() EntityId
    Name() string
    GetState(id StateId) State
    SetState(state State)
    AddAction(action Action)
    RemoveAction(action Action)
}

// Simplified declaration
type StateList map[StateId]State
// Simplified declaration
type ActionList map[ActionId]Action

// Contains all the data that each component needs.
// TODO: Rename to 'Component'?
type CmpData struct {
    // Use maps for easy/add remove for now
    states  StateList
    actions ActionList
    input   chan Msg
}

// Creates a CmpData and initializes its containers.
func NewCmpData() *CmpData {
    states := make(StateList)
    actions := make(ActionList)
    return &CmpData{states, actions, nil}
}

// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Id() EntityId { return 0 }
// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Name() string { return "CmpData" }

// The next functions form the core functionality of a component.

// Returns the requested State. It is up to the caller to verify that the wanted
// state was actually returned.
func (cd *CmpData) GetState(id StateId) State {
    return cd.states[id]
}

// Set the value of the passed State. Replaces any existing State that is the same.
func (cd *CmpData) SetState(state State) {
    cd.states[state.Id()] = state
}

// Adds to an Entity's actions, causing the Action to be executed on the next tick.
func (cd *CmpData) AddAction(action Action) {
    cd.actions[action.Id()] = action
}

// Removes the Action from an Entity's actions.
func (cd *CmpData) RemoveAction(action Action) {
    cd.actions[action.Id()] = nil, false
}

// Main loop which handles all component tasks.
func (cd *CmpData) Run(input chan Msg) {
    cd.input = input

    for {
        msg := <-input

        // Call the appropriate function based on the msg type
        switch m := msg.(type) {
        case MsgTick:
            cd.update()
            m.Origin <- MsgTick{input} // Reply that we are updated
        case MsgGetState:
            cd.sendState(m)
        case MsgGetAllStates:
            cd.sendAllStates(m)
        case MsgAddAction:
            cd.AddAction(m.Action)
        }
    }
}

// Loop through each Action and let it run
func (cd *CmpData) update() {
    for _, v := range cd.actions {
        v.Act(cd)
    }
}

// Send back the requested State on the provided channel
func (cd *CmpData) sendState(msg MsgGetState) {
    state := cd.states[msg.Id]
    msg.StateReply <- state
}

// Send back all states on the provided channel
func (cd *CmpData) sendAllStates(msg MsgGetAllStates) {
    for _, v := range cd.states {
        msg.StateReply <- v
    }
    close(msg.StateReply)
}
