// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package comm

import (
    "protocol"
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

func makeAssignControl(uid int32, revoked bool) (msg *protocol.Message) {
    ctrl := &protocol.AssignControl{
        Uid:     &uid,
        Revoked: &revoked,
    }

    return &protocol.Message{
        AssignControl: ctrl,
        Type:          protocol.NewMessage_Type(protocol.Message_ASSIGNCONTROL),
    }
}

func makeEntityDeath(uid int32, name string) (msg *protocol.Message) {
    entityDeath := &protocol.EntityDeath{
        Uid:    &uid,
        Name:   &name,
    }

    return &protocol.Message{
        EntityDeath: entityDeath,
        Type:        protocol.NewMessage_Type(protocol.Message_ENTITYDEATH),
    }
}
