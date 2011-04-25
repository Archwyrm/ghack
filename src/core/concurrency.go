// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Concurrent communication helpers

package core

// Implementing this interface allows a type to safely send and receive
// messages without risking a deadlock.
type MsgHandler interface {
    HandleMsg(msg Msg)
    Chan() chan Msg
}

// Send and receive simultaneously to prevent deadlock. Will continue to
// receive new messages until the send goes through. hnd will be used to handle
// any incoming messages on its input channel.
func Send(handler MsgHandler, out chan Msg, msg Msg) {
    for {
        select {
        case out <- msg:
            return
        case m := <-handler.Chan():
            handler.HandleMsg(m)
        }
    }
}

// Receive on the given channel and receive on the main channel simultaneously
// to prevent deadlock. Will continue to receive new messages until the receive
// to in goes through. hnd will be used to handle any incoming messages on its
// input channel.
func Recv(hnd MsgHandler, in chan Msg) Msg {
    for {
        select {
        case m := <-in:
            return m
        case m := <-hnd.Chan():
            hnd.HandleMsg(m)
        }
    }
    return nil // We will never hit this point
}

// HandlerQueue is a helper which queues messages and returns them when asked
// for. It partly implements the MsgHandler interface, thus is intended for
// embedding in a struct that wishes to implement the interface.
type HandlerQueue struct {
    msgs []Msg
}

func NewHandlerQueue() *HandlerQueue {
    hq := &HandlerQueue{make([]Msg, 0, 2)}
    return hq
}

// HandleMsg handles a message by queueing it for later retrieval.
func (hq *HandlerQueue) HandleMsg(msg Msg) {
    hq.msgs = append(hq.msgs, msg)
}

// GetMsg gets a message out of the queue or if there are none, listens on the
// passed channel for the next one.
func (hq *HandlerQueue) GetMsg(ch chan Msg) Msg {
    if len(hq.msgs) > 0 {
        next := hq.msgs[0]
        hq.msgs = hq.msgs[1:len(hq.msgs)]
        return next
    }
    return <-ch
}
