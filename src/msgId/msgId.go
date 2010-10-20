// Namespace for message identification
package msgId

const (
    MsgTick      = iota // Signal to the component that it should update
    MsgGetState         // Request a State to be returned on StateReply chan
    MsgAddAction        // Add some kind of action to the Entity's list
)
