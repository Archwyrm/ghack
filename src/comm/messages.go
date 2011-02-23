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
func packState(state core.State) *protocol.StateValue {
    msg := &protocol.StateValue{} // Init message
    t := reflect.NewValue(state)
    state_v, ok := t.(*reflect.StructValue)
    if !ok {
        panic("State is non-struct type!")
    }
    if state_v.NumField() > 1 {
        panic("Protocol only supports states with one field currently")
    }
    // TODO: Add support for state structs with more than one field
    val := state_v.Field(0)
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
        //TODO: Implement repeated values
    default:
        panic("State value not supported:" + val.Type().String())
    }
    return msg
}
