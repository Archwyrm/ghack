// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Tests that reflection based processing on various types of State creates
// the right messages to send to clients.

package comm

import (
    "testing"
    "reflect"
    .   "core"
    "protocol"
    "github.com/tm1rbrt/s3dm"
    "goprotobuf.googlecode.com/hg/proto"
)

type multipleFieldState struct {
    Value1 int
    Value2 int
}

func (x multipleFieldState) Id() StateId  { return 1 }
func (x multipleFieldState) Name() string { return "MultipleFieldState" }

type sliceFieldState struct {
    Value []int
}

func (x sliceFieldState) Id() StateId  { return 2 }
func (x sliceFieldState) Name() string { return "SliceFieldState" }

type ptrFieldState struct {
    Value *int
}

func (x ptrFieldState) Id() StateId  { return 3 }
func (x ptrFieldState) Name() string { return "ptrFieldState" }

type v3FieldState struct {
    Value s3dm.V3
}

func (x v3FieldState) Id() StateId  { return 4 }
func (x v3FieldState) Name() string { return "v3FieldState" }

func TestSingleFieldState(t *testing.T) {
    // testState is made available by observer_test.go
    state := testState{9}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_INT)
    sv.IntVal = proto.Int(state.Value)

    equalOrError(t, sv, packState(state))
}

func TestMultipleFieldState(t *testing.T) {
    a, b := 1, 2
    state := multipleFieldState{a, b}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_ARRAY)

    inner_v1 := makeIntValMsg(a)
    inner_v2 := makeIntValMsg(b)
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

func TestV3FieldState(t *testing.T) {
    vec := s3dm.V3{9, 9, 9}
    state := v3FieldState{vec}
    sv := &protocol.StateValue{}
    sv.Type = protocol.NewStateValue_Type(protocol.StateValue_VECTOR3)
    sv.Vector3Val = &protocol.Vector3{&vec.X, &vec.Y, &vec.Z, nil}

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
