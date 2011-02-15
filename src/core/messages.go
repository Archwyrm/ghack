// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package core

// Universal message interface for components and services.
// Does not currently do anything special, but is reserved for any possible
// future use and thus should be specified where any other actual message
// types are expected.
type Msg interface{}

// Message to update
type MsgTick struct{}

// Message requesting a certain state to be returned
// Contains a channel where the reply should be sent
type MsgGetState struct {
    Id         StateId
    StateReply chan State
}

// Message requesting all states that an entity has
// StateReply will be ranged over by the originator of request
type MsgGetAllStates struct {
    StateReply chan State
}

// Message to add an action that contains the action to be added
type MsgAddAction struct {
    Action Action
}

// Requests and returns a list of entity handles (channels) and types.
// Message with an empty list is considered a request, while non-empty contains
// an actual list. Each pair has matching indices in their respective slices.
type MsgListEntities struct {
    Reply    chan Msg // Channel to reply on
    Entities []chan Msg
    Types    []EntityId
}
