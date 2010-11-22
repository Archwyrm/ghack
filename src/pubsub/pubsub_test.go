package pubsub_test

import (
    "testing"
    "testing/script"
    "fmt"
    "pubsub/pubsub"
    "core/core"
)

var testData = []string{
    "testing testing, one two three",
    "test here, test there, test everywhere",
    "quit being so testy!",
}

// Tests subscription followed by publishing
func TestSubscribeAndPublish(t *testing.T) {
    // Initialize
    topic := "test"
    psObj := pubsub.NewPubSub()
    ps := make(chan core.ServiceMsg)
    go psObj.Run(ps)

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

// Relay proper messages to pubsub for testing purposes
func startRelay(ps chan core.ServiceMsg, topic string) (relay chan interface{}) {
    relay = make(chan interface{})
    go func() {
        for {
            ps <- pubsub.PublishMsg{topic, <-relay}
        }
    }()
    return
}

// Makes 'count' channels and subscribes them
func makeAndSubscribe(ps chan core.ServiceMsg, topic string, count int) (chans []chan interface{}) {
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
