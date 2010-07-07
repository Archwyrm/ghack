package main

// Struct for requesting a component to do something
// Different fields will be filled in depending on the type of message
type CmpMsg struct {
    Id         int
    StateId    string
    StateReply chan State
    Action     Action
}

const (
    MsgTick      = iota // Signal to the component that it should update
    MsgGetState         // Request a State to be returned on StateReply chan
    MsgAddAction        // Add some kind of action to the Entity's list
)
