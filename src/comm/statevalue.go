// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "reflect"
    .   "core"
    "protocol"
    "goprotobuf.googlecode.com/hg/proto"
)

// Creates the right type of StateValue message for an arbitrary State type.
func packState(state State) (msg *protocol.StateValue) {
    val := reflect.NewValue(state)
    state_v, ok := val.(*reflect.StructValue)
    if !ok {
        panic("State is non-struct type!")
    }

    field_num := state_v.NumField()
    if field_num > 1 { // If we have multiple fields, treat as array
        msg = &protocol.StateValue{}
        msg.Type = protocol.NewStateValue_Type(protocol.StateValue_ARRAY)
        msg.ArrayVal = makeStateValueArray(state_v, field_num)
    } else { // Single field
        msg = readField(state_v.Field(0))
    }
    return msg
}

// Reads a single arbitrary type and returns the proper StateValue.
func readField(val reflect.Value) *protocol.StateValue {
    msg := &protocol.StateValue{}
    switch f := val.(type) {
    case *reflect.BoolValue:
        msg.Type = protocol.NewStateValue_Type(protocol.StateValue_BOOL)
        msg.BoolVal = proto.Bool(f.Get())
    case *reflect.IntValue:
        msg.Type = protocol.NewStateValue_Type(protocol.StateValue_INT)
        msg.IntVal = proto.Int(int(f.Get()))
    case *reflect.FloatValue:
        msg.Type = protocol.NewStateValue_Type(protocol.StateValue_FLOAT)
        msg.FloatVal = proto.Float32(float32(f.Get()))
    case *reflect.StringValue:
        msg.Type = protocol.NewStateValue_Type(protocol.StateValue_STRING)
        msg.StringVal = proto.String(f.Get())
    case *reflect.SliceValue:
        msg.Type = protocol.NewStateValue_Type(protocol.StateValue_ARRAY)
        msg.ArrayVal = makeStateValueArray(f, f.Len())
    case *reflect.StructValue:
        msg = readStructField(f)
    case *reflect.PtrValue:
        return readField(reflect.Indirect(f)) // Dereference and recurse
    default:
        panic("State value not supported: " + val.Type().String())
    }
    return msg
}

// Reads a field that is not a builtin type, but a user created struct. These
// types have specific message types so the receiving end can identify them.
func readStructField(val *reflect.StructValue) *protocol.StateValue {
    t := val.Type()
    switch t.Name() {
    case "V3": // Vector type:
        return makeVector3(val)
    }
    panic("Struct value not supported: " + t.String())
    return nil // Will never get here
}

// Creates a slice of StateValues based on multiple arbitrary types.
func makeStateValueArray(value reflect.Value, num int) []*protocol.StateValue {
    msg_array := make([]*protocol.StateValue, 0, num)

    // One of these will be valid, the other will not
    struct_v, ok := value.(*reflect.StructValue)
    slice_v, ok2 := value.(*reflect.SliceValue)
    _, _ = ok, ok2 // These can be safely ignored

    // Loop through the struct's fields
    for i := 0; i < num; i++ {
        var val reflect.Value
        if struct_v != nil {
            val = struct_v.Field(i)
        } else {
            val = slice_v.Elem(i)
        }
        msg_array = append(msg_array, readField(val))
    }
    return msg_array
}

// Makes a Vector3 StateValue. Panics if the StructValue fields do not match the vector.
func makeVector3(v *reflect.StructValue) *protocol.StateValue {
    // If we panic here, struct layout was not as expected
    x := v.FieldByName("X").(*reflect.FloatValue).Get()
    y := v.FieldByName("Y").(*reflect.FloatValue).Get()
    z := v.FieldByName("Z").(*reflect.FloatValue).Get()

    vector3 := &protocol.Vector3{&x, &y, &z, nil}
    sv := &protocol.StateValue{
        Type:       protocol.NewStateValue_Type(protocol.StateValue_VECTOR3),
        Vector3Val: vector3,
    }
    return sv
}
