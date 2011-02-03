// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package pubsub_test

import (
    "testing"
    "testing/script"
    "fmt"
    "time"
    "pubsub/pubsub"
    "core/core"
)

var testData = []string{
    "testing testing, one two three",
    "test here, test there, test everywhere",
    "quit being so testy!",
}

var topic = "test"

// Tests subscription followed by publishing
func TestSubscribeAndPublish(t *testing.T) {
    // Initialize
    ps := startPubSub()

    // Subscribe
    chans := makeAndSubscribe(ps, topic, 10)

    relay := startRelay(ps, topic)

    // Publish and verify one message at a time
    var events, recvs []*script.Event
    for _, data := range testData {
        send := script.NewEvent("send", recvs, script.Send{relay, data})
        recvs = makeRecvEvents(chans, []*script.Event{send}, data)
        events = append(append(events, send), recvs...)
    }
    err := script.Perform(0, events)

    if err != nil {
        t.Errorf("Sent and published values do not match!\n%s", err.String())
    }
}

// Test subscribing and then removing a subscription
func TestUnsubscribe(t *testing.T) {
    ps := startPubSub()
    chans := makeAndSubscribe(ps, topic, 5)

    ch_i := 2
    ps <- pubsub.UnsubscribeMsg{topic, chans[ch_i]}

    quit := make(chan bool)

    // Error if we receive anything on the unsubscribed channel
    go func() {
        select {
        case <-chans[ch_i]:
            t.Fatalf("Received message on unsubscribed channel!")
        case <-quit:
            return
        }
    }()

    // Drain the subscribed channels
    for i, ch := range chans {
        if i == ch_i {
            continue
        }
        go func(ch chan interface{}) {
            <-ch
        }(ch)
    }

    ps <- pubsub.PublishMsg{topic, testData[0]}
    time.Sleep(500 * 10e5) // Wait 500ms
    quit <- true
}

// Relay proper messages to pubsub for testing purposes
func startRelay(ps chan core.Msg, topic string) (relay chan interface{}) {
    relay = make(chan interface{})
    go func() {
        for {
            ps <- pubsub.PublishMsg{topic, <-relay}
        }
    }()
    return
}

// Makes 'count' channels and subscribes them
func makeAndSubscribe(ps chan core.Msg, topic string, count int) (chans []chan interface{}) {
    for i := 0; i < count; i++ {
        ch := make(chan interface{})
        ps <- pubsub.SubscribeMsg{topic, ch}
        chans = append(chans, ch)
    }
    return
}

// Makes receive events from an array of channels
// Returns a ready list of events
func makeRecvEvents(chans []chan interface{}, pre []*script.Event, data interface{}) (events []*script.Event) {
    for i, ch := range chans {
        name := fmt.Sprintf("recv %d", i)
        ev := script.NewEvent(name, pre, script.Recv{ch, data})
        events = append(events, ev)
    }
    return
}

// Starts the PubSub in a goroutine and returns a channel to it
func startPubSub() (ps chan core.Msg) {
    psObj := pubsub.NewPubSub()
    ps = make(chan core.Msg)
    go psObj.Run(ps)
    return
}
