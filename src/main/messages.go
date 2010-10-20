package main

import "msgId/msgId"

// Interface for requesting a component to do something
type CmpMsg interface {
    Id() int
}

// Message to update
type MsgTick struct{}

func (msg MsgTick) Id() int { return msgId.MsgTick }

// Message requesting a certain state to be returned
// Contains a channel where the reply should be sent
type MsgGetState struct {
    StateId    string
    StateReply chan State
}

func (msg MsgGetState) Id() int { return msgId.MsgGetState }

// Message to add an action that contains the action to be added
type MsgAddAction struct {
    Action Action
}

func (msg MsgAddAction) Id() int { return msgId.MsgAddAction }
