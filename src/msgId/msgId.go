// Namespace for message identification
package msgId

type MsgId int

const (
    // Component messages
    Tick      = iota // Signal to the component that it should update
    GetState         // Request a State to be returned on StateReply chan
    AddAction        // Add some kind of action to the Entity's list

    // PubSub messages
    Publish     // Publish a message to a topic
    Subscribe   // Subscribe to a topic
    Unsubscribe // Unsubscribe from a topic
)
