// Namespace for message identification
package msgId

type MsgId int

const (
    // Component messages
    MsgTick      = iota // Signal to the component that it should update
    MsgGetState         // Request a State to be returned on StateReply chan
    MsgAddAction        // Add some kind of action to the Entity's list

    // PubSub messages
    Publish     // Publish a message to a topic
    Subscribe   // Subscribe to a topic
    Unsubscribe // Unsubscribe from a topic
)
