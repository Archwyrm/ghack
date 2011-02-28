// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "reflect"
    "core/core"
    "protocol/protocol"
    "goprotobuf.googlecode.com/hg/proto"
)

func makeConnect() (msg *protocol.Message) {
    connect := &protocol.Connect{Version: proto.Uint32(ProtocolVersion)}

    return &protocol.Message{
        Connect: connect,
        Type:    protocol.NewMessage_Type(protocol.Message_CONNECT),
    }
}


func makeDisconnect(reason int32, reasonString string) (msg *protocol.Message) {
    disconnect := &protocol.Disconnect{
        Reason:    protocol.NewDisconnect_Reason(reason),
        ReasonStr: proto.String(reasonString),
    }

    return &protocol.Message{
        Disconnect: disconnect,
        Type:       protocol.NewMessage_Type(protocol.Message_DISCONNECT),
    }
}

func makeLogin(name string, authToken string, permissions uint32) (msg *protocol.Message) {
    login := &protocol.Login{
        Name:        proto.String(name),
        Authtoken:   proto.String(authToken),
        Permissions: proto.Uint32(permissions),
    }

    return &protocol.Message{
        Login: login,
        Type:  protocol.NewMessage_Type(protocol.Message_LOGIN),
    }
}

func makeLoginResult(succeeded bool, reason int32) (msg *protocol.Message) {
    loginResult := &protocol.LoginResult{
        Succeeded: proto.Bool(succeeded),
        Reason:    protocol.NewLoginResult_Reason(reason),
    }

    return &protocol.Message{
        LoginResult: loginResult,
        Type:        protocol.NewMessage_Type(protocol.Message_LOGINRESULT),
    }
}

func makeAddEntity(id int32, name string) (msg *protocol.Message) {
    addEntity := &protocol.AddEntity{
        Id:   proto.Int32(id),
        Name: proto.String(name),
    }

    return &protocol.Message{
        AddEntity: addEntity,
        Type:      protocol.NewMessage_Type(protocol.Message_ADDENTITY),
    }
}

func makeRemoveEntity(id int32, name string) (msg *protocol.Message) {
    removeEntity := &protocol.RemoveEntity{
        Id:   proto.Int32(id),
        Name: proto.String(name),
    }

    return &protocol.Message{
        RemoveEntity: removeEntity,
        Type:         protocol.NewMessage_Type(protocol.Message_REMOVEENTITY),
    }
}

func makeUpdateState(id int32, stateId string, value *protocol.StateValue) (msg *protocol.Message) {
    updateState := &protocol.UpdateState{
        Id:      proto.Int32(id),
        StateId: proto.String(stateId),
        Value:   value,
    }

    return &protocol.Message{
        UpdateState: updateState,
        Type:        protocol.NewMessage_Type(protocol.Message_UPDATESTATE),
    }
}

// Creates the right type of StateValue message for an arbitrary State type.
func packState(state core.State) (msg *protocol.StateValue) {
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
    default:
        panic("State value not supported:" + val.Type().String())
    }
    return msg
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
