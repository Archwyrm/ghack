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
// Identifies an individual entity instance
type UniqueId int

// Holds some kind of state data for a particular named property of an Entity.
type State interface {
    // Returns a unique ID for each Action (defined in cmpId package)
    Id() StateId
    // Returns a unique and semi-descriptive name for each Action (defined by
    // the State)
    Name() string
}

// The Action enacts changes in Entity state. It may be considered a
// transactional state change, or since it arrives by message, a closure.
type Action interface {
    // Returns a unique ID for each Action (defined in cmpId package)
    Id() ActionId
    // Returns a unique and semi-descriptive name for each Action (defined by
    // the Action)
    Name() string
    Act(ent Entity, svc ServiceContext)
}

// An Entity is a struct composed from various States and Actions, which each
// make up its data and functionality respectively.
type Entity interface {
    // Returns unique ID for the *instance* of this entity
    Uid() UniqueId
    // Returns a unique entity type ID (defined in cmpId package)
    Id() EntityId
    // Name() returns a unique and semi-descriptive name for each Entity (defined
    // by the Entity)
    Name() string
    // Returns the requested State by ID or nil if it does not exist
    GetState(id StateId) State
    // Sets the value of the passed State within the Entity. Overwrites any
    // previous state with the same ID.
    SetState(state State)
    // Adds the Action to the Entity.
    AddAction(action Action)
    // Removes the Action from the Entity.
    RemoveAction(action Action)
    // Runs the Entity's main loop.
    Run(svc ServiceContext)
    // Returns the Entity's communication channel. Returns nil if Run() has not
    // yet been called.
    Chan() chan Msg
}

// Simplified declaration
type StateList map[StateId]State
// Simplified declaration
type ActionList map[ActionId]Action

// Contains all the data that each component needs.
// TODO: Rename to 'Component'?
type CmpData struct {
    uid UniqueId
    // Use maps for easy/add remove for now
    states  StateList
    actions ActionList
    input   chan Msg
}

// Creates a CmpData and initializes its containers.
func NewCmpData(uid UniqueId) *CmpData {
    states := make(StateList)
    actions := make(ActionList)
    ch := make(chan Msg)
    return &CmpData{uid, states, actions, ch}
}

// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Id() EntityId { return 0 }
// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Name() string { return "CmpData" }

// The next functions form the core functionality of a component.

func (cd *CmpData) Uid() UniqueId { return cd.uid }

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

// Returns the input channel.
func (cd *CmpData) Chan() chan Msg { return cd.input }

// Main loop which handles all component tasks.
func (cd *CmpData) Run(svc ServiceContext) {
    for {
        msg := <-cd.input

        // Call the appropriate function based on the msg type
        switch m := msg.(type) {
        case MsgTick:
            cd.update(svc)
            m.Origin <- MsgTick{cd.input} // Reply that we are updated
        case MsgGetState:
            cd.sendState(m)
        case MsgGetAllStates:
            cd.sendAllStates(m)
        case MsgSetState:
            cd.SetState(m.State)
        case MsgAddAction:
            cd.AddAction(m.Action)
        case MsgRunAction:
            m.Action.Act(cd, svc)
            if m.Add {
                cd.AddAction(m.Action)
            }
        }
    }
}

// Loop through each Action and let it run
func (cd *CmpData) update(svc ServiceContext) {
    for _, v := range cd.actions {
        v.Act(cd, svc)
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

// Entity descriptor, contains all the relevant information for a given entity
// in one neat little package.
type EntityDesc struct {
    Chan chan Msg // Channel to entity
    Uid  UniqueId // Unique id
    Id   EntityId // Type id
    Name string   // Name of entity
}

// Returns a new entity descriptor based off a given entity and channel.
func NewEntityDesc(ent Entity) *EntityDesc {
    return &EntityDesc{ent.Chan(), ent.Uid(), ent.Id(), ent.Name()}
}
