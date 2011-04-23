// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package core

// Universal message interface for components and services.
// Does not currently do anything special, but is reserved for any possible
// future use and thus should be specified where any other actual message
// types are expected.
type Msg interface{}

// Message to signal an update and/or updated status.
// A completion of update reply should be sent to the Origin channel.
type MsgTick struct {
    Origin chan Msg // Identifies the sources of the tick
}

// Tells the receiver to quit, shutdown, stop, halt, cease operations, close for
// business, etc..
type MsgQuit struct{}

// Message requesting a certain state to be returned
// Contains a channel where the reply should be sent
type MsgGetState struct {
    Id         StateId
    StateReply chan State
}

// Message requesting that a certain state should be set
type MsgSetState struct {
    State State
}

// Message requesting all states that an entity has
// StateReply will be ranged over by the originator of request and should be
// closed once all states have been sent.
type MsgGetAllStates struct {
    StateReply chan State
}

// Message to add an action that contains the action to be added
type MsgAddAction struct {
    Action Action
}

// Requests that the action be run immediately. Will be optionally added for
// re-use depending on the Add variable.
type MsgRunAction struct {
    Action Action
    Add    bool
}

// Requests and returns a list of entity handles (channels), types, and names.
// Message with an empty list is considered a request, while non-empty contains
// an actual list. Each triple of lists has matching indices in their respective
// slices.
type MsgListEntities struct {
    Reply    chan Msg // Channel to reply on
    Entities []*EntityDesc
}

// Signals that a specific entity has been added to the game
type MsgEntityAdded struct {
    Entity *EntityDesc
}

// Signals that a specific entity has been removed from the game
type MsgEntityRemoved struct {
    Entity *EntityDesc
}

// Assigns control of this entity to a client
type MsgAssignControl struct {
    Uid     UniqueId // Entity to be given to a client
    Revoked bool     // True, if control is to be removed
}

// Signifies that an entity died
type MsgEntityDeath struct {
    Entity *EntityDesc
}

// Represents damage dealt in combat
type MsgCombatHit struct {
    Attacker *EntityDesc
    Victim   *EntityDesc
    Damage   float32
}
