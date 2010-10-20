// Contains everything necessary for component system

package main

import "msgId/msgId"

// Holds some kind of state data for a particular named property of an Entity.
// Id() returns a unique ID for each Action (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Action (defined
// by each Entity)
type State interface {
    Id() int
    Name() string
}

// The Action enacts changes in Entity state. It may be considered a
// transactional state change, or since it arrives by message, a closure.
// Id() returns a unique ID for each Action (defined in cmpId package)
// Name() returns a unique and semi-descriptive name for each Action (defined
// by each Entity)
type Action interface {
    Id() int
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
    Id() int
    Name() string
    GetState(state State) State
    SetState(state State)
    AddAction(action Action)
    RemoveAction(action Action)
}

// Contains all the data that each component needs.
// TODO: Rename to 'Component'?
type CmpData struct {
    // Use maps for easy/add remove for now
    states  StateList
    actions ActionList
    input   chan CmpMsg
}

// Simplified declaration
type StateList map[string]State
// Simplified declaration
type ActionList map[string]Action

// Creates a CmpData and initializes its containers.
func NewCmpData() *CmpData {
    states := make(StateList)
    actions := make(ActionList)
    return &CmpData{states, actions, nil}
}

// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Id() int { return 0 }
// Added to satisfy the Entity interface, clobbered by embedding.
func (cd CmpData) Name() string { return "CmpData" }

// The next functions form the core functionality of a component.

// Returns the requested State. TODO: Take StateId?
func (cd CmpData) GetState(state State) State {
    ret := cd.states[state.Name()]
    return ret
}

// Set the value of the passed State. Replaces any existing State that is the same.
func (cd CmpData) SetState(state State) {
    cd.states[state.Name()] = state
}

// Adds to an Entity's actions, causing the Action to be executed on the next tick.
func (cd CmpData) AddAction(action Action) {
    cd.actions[action.Name()] = action
}

// Removes the Action from an Entity's actions.
func (cd CmpData) RemoveAction(action Action) {
    cd.actions[action.Name()] = nil, false
}

// Main loop which handles all component tasks.
func (cd CmpData) Run(input chan CmpMsg) {
    cd.input = input

    for {
        msg := <-input

        // Call the appropriate function based on the msg type
        switch {
        case msg.Id() == msgId.MsgTick:
            cd.update()

        case msg.Id() == msgId.MsgGetState:
            m, ok := msg.(MsgGetState)
            if ok {
                cd.sendState(m)
            }
        case msg.Id() == msgId.MsgAddAction:
            m, ok := msg.(MsgAddAction)
            if ok {
                cd.AddAction(m.Action)
            }
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
