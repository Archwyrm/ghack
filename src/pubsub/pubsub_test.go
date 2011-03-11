// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package pubsub_test

import (
    "testing"
    "time"
    "pubsub"
    "core"
)

var testData = []string{
    "testing testing, one two three",
    "test here, test there, test everywhere",
    "quit being so testy!",
}

var topic = "test"

func TestSubscribeAndPublish(t *testing.T) {
    ps := startPubSub()
    chans := makeAndSubscribe(ps, topic, 3)
    verified := make(chan bool)

    // Publish and verify one message at a time
    for _, msg := range testData {
        // Publish message
        go func(ps chan core.Msg, data string) {
            ps <- pubsub.PublishMsg{topic, data}
        }(ps, msg)

        // Receive and verify
        go func() {
            for {
                select {
                case iface := <-chans[0]:
                    verify(t, msg, iface, verified)
                case iface := <-chans[1]:
                    verify(t, msg, iface, verified)
                case iface := <-chans[2]:
                    verify(t, msg, iface, verified)
                }
            }
        }()

        // Once a value has been sent len(chans) times, all chans are verified
        for _ = range chans {
            <-verified
        }
    }
}

func verify(t *testing.T, msg string, iface interface{}, verified chan bool) {
    str := iface.(string)
    if str == msg {
        verified <- true
    } else {
        t.Fatalf("Sent message does not match received message, wanted: '%s'\ngot: '%s'", msg, str)
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
        go func(ch pubsub.ChanType) {
            <-ch
        }(ch)
    }

    ps <- pubsub.PublishMsg{topic, testData[0]}
    time.Sleep(500 * 10e5) // Wait 500ms
    quit <- true
}

// Makes 'count' channels and subscribes them
func makeAndSubscribe(ps chan core.Msg, topic string, count int) (chans []pubsub.ChanType) {
    for i := 0; i < count; i++ {
        ch := make(pubsub.ChanType)
        ps <- pubsub.SubscribeMsg{topic, ch}
        chans = append(chans, ch)
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
