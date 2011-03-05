// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Tests that reflection based processing on various types of core.State creates
// the right messages to send to clients.

package comm

import (
    "testing"
    "reflect"
    "core/core"
    "protocol/protocol"
    "goprotobuf.googlecode.com/hg/proto"
)

type multipleFieldState struct {
    Value1 int
    Value2 int
}

func (x multipleFieldState) Id() core.StateId { return 1 }
func (x multipleFieldState) Name() string     { return "MultipleFieldState" }

type sliceFieldState struct {
    Value []int
}

func (x sliceFieldState) Id() core.StateId { return 2 }
func (x sliceFieldState) Name() string     { return "SliceFieldState" }
func newSliceFieldState(s []int) *sliceFieldState { return &sliceFieldState{s}}

type ptrFieldState struct {
    Value *int
}

func (x ptrFieldState) Id() core.StateId { return 3 }
func (x ptrFieldState) Name() string     { return "ptrFieldState" }

func TestSingleFieldState(t *testing.T) {
    // testState is made available by observer_test.go
    state := testState{9}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_INT)
    sv.IntVal = proto.Int(state.Value)

    equalOrError(t, sv, packState(state))
}

func TestMultipleFieldState(t *testing.T) {
    state := multipleFieldState{1, 2}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_ARRAY)

    inner_v1 := makeIntValMsg(1)
    inner_v2 := makeIntValMsg(2)
    sv.ArrayVal = append([]*protocol.StateValue{}, inner_v1, inner_v2)

    equalOrError(t, sv, packState(state))
}

func TestSliceFieldState(t *testing.T) {
    s := []int{1, 2, 3, 4, 5}
    state := sliceFieldState{s}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_ARRAY)

    msgs := make([]*protocol.StateValue, 0, len(s))
    for _, i := range s {
        msgs = append(msgs, makeIntValMsg(i))
    }
    sv.ArrayVal = msgs

    equalOrError(t, sv, packState(state))
}

func TestPtrFieldState(t *testing.T) {
    num := 9
    state := ptrFieldState{&num}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_INT)
    sv.IntVal = proto.Int(*state.Value)

    equalOrError(t, sv, packState(state))
}

// Returns an error if two StateValues are not equal
func equalOrError(t *testing.T, sv, pack interface{}) {
    if !reflect.DeepEqual(sv, pack) {
        t.Error("Test state and packed state value messages do not match!")
    }
}

// Returns a StateValue with IntVal set to the passed value x
func makeIntValMsg(x int) *protocol.StateValue {
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_INT)
    sv.IntVal = proto.Int(x)
    return sv
}
