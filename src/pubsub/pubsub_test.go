package pubsub_test

import (
    "testing"
    "pubsub/pubsub"
    "core/core"
)

type subscriber struct {
    received bool       // Whether the message has been received or not
    ch chan interface{} // Subscription channel
}

var subscribers = []*subscriber {
    &subscriber{false, make(chan interface{})},
    &subscriber{false, make(chan interface{})},
    &subscriber{false, make(chan interface{})},
}

var testData = []string {
    "testing testing, one two three",
    "test here, test there, test everywhere",
    "quit being so testy!",
}

var topic = "test"
var verified = false

func TestPubSub(t *testing.T) {
    // Initialize
    psObj := pubsub.NewPubSub()
    ps := make(chan core.ServiceMsg)
    go psObj.Run(ps)

    // Subscribe
    for _, s := range subscribers {
        ps <- pubsub.SubscribeMsg{topic, s.ch}
    }

    // Publish and verify one message at a time
    for _, msg := range testData {
        // Publish message
        go func(ps chan core.ServiceMsg, data string) {
            ps <- pubsub.PublishMsg{topic, data}
        } (ps, msg)

        // Receive and verify
        var iface interface{}
        for !verified {
            select {
            case iface = <-subscribers[0].ch:
                verify(t, subscribers[0], msg, iface)
            case iface = <-subscribers[1].ch:
                verify(t, subscribers[1], msg, iface)
            case iface = <-subscribers[2].ch:
                verify(t, subscribers[2], msg, iface)
            }
        }
        reset() // Reset verification for next message
    }
}

func verify(t *testing.T, sub *subscriber, msg string, iface interface{}) {
    str := iface.(string)
    if str == msg {
        for i, s := range subscribers {
            if sub == s {
                subscribers[i].received = true
            }
        }
    } else {
        t.Fatalf("Sent message does not match received message, wanted: '%s'\ngot: '%s'", msg, str)
    }

    // Check whether all messages have been received
    received := true
    for _, s := range subscribers {
        received = received && s.received
    }
    verified = received
}

func reset() {
    for i, _ := range subscribers {
        subscribers[i].received = false
    }
    verified = false
}
