// Implements the publish/subscribe messaging model.
//
// This model allows asynchronus communications where senders and receivers
// are not explicit. Rather, users subscribe to certain named topics and when
// a message is published to a topic, all the subscribers receive this message.
package pubsub

import (
    "core/core"
    "msgId/msgId"
)

// Message to signal publishing of the passed data
type PublishMsg struct {
    Topic string
    Data  interface{}
}

func (x PublishMsg) Id() msgId.MsgId { return msgId.Publish }

// Message to setup a subscription to a given topic
type SubscribeMsg struct {
    Topic     string
    ReplyChan chan interface{}
}

func (x SubscribeMsg) Id() msgId.MsgId { return msgId.Subscribe }

// Publish/Subscribe struct
type PubSub struct {
    subscriptions map[string][]chan interface{}
}

// Creates a new PubSub and returns a pointer to it
func NewPubSub() *PubSub {
    return &PubSub{make(map[string][]chan interface{})}
}

// Starts a loop to receive and handle messages from the passed channel
func (ps *PubSub) Run(input chan core.ServiceMsg) {
    for {
        msg := <-input

        switch {
        case msg.Id() == msgId.Publish:
            ps.publish(msg.(PublishMsg))

        case msg.Id() == msgId.Subscribe:
            ps.subscribe(msg.(SubscribeMsg))
        }
    }
}

// Sends a message to subscribers
func (ps *PubSub) publish(msg PublishMsg) {
    for _, sub := range ps.subscriptions[msg.Topic] {
        sub <- msg.Data
    }
}

// Adds a subscription to the appropriate topic
func (ps *PubSub) subscribe(msg SubscribeMsg) {
    subscribers := ps.subscriptions[msg.Topic]
    ps.subscriptions[msg.Topic] = append(subscribers, msg.ReplyChan)
}
