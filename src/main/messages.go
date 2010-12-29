// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package main

import (
    "cmpId/cmpId"
    "msgId/msgId"
)

// Interface for requesting a component to do something
type CmpMsg interface {
    Id() int
}

// Message to update
type MsgTick struct{}

func (msg MsgTick) Id() int { return msgId.Tick }

// Message requesting a certain state to be returned
// Contains a channel where the reply should be sent
type MsgGetState struct {
    StateId    cmpId.StateId
    StateReply chan State
}

func (msg MsgGetState) Id() int { return msgId.GetState }

// Message to add an action that contains the action to be added
type MsgAddAction struct {
    Action Action
}

func (msg MsgAddAction) Id() int { return msgId.AddAction }
