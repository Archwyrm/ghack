// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package sf

import (
    "log"
    "github.com/tm1rbrt/s3dm"
    .   "core"
    "game"
    "protocol"
)

// An avatar is the agent of a client that acts on its behalf dealing with the
// entity system. It is somewhat the opposite of observer (in the comm package),
// as observer sends messages to the client, avatar receives messages from the
// client. While observer has no knowledge of the game, avatar does. It knows
// what it is controlling and how to control it. Thus a game must provide an
// avatar implementation in order to allow clients to interact with the game.
type avatar struct {
    svc    ServiceContext
    player EntityDesc
}

// Starts an avatar on behalf of a connected client. Takes a current ServiceContext
// and channel of received messages. Returns a control channel for the avatar
// and the uid of the entity created for the client.
func MakeAvatar(svc ServiceContext, input chan *protocol.Message) (chan Msg,
UniqueId) {
    ctrl := make(chan Msg)
    reply := make(chan *EntityDesc)
    svc.Game <- game.MsgSpawnEntity{InitPlayer, reply}
    player := <-reply
    a := avatar{svc, *player}
    go a.control(ctrl, input)
    return ctrl, player.Uid
}

func (a *avatar) control(ctrl <-chan Msg, input <-chan *protocol.Message) {
    for {
        select {
        case msg := <-ctrl:
            switch m := msg.(type) {
            // TODO: Handle MsgTick?
            case MsgQuit:
                a.player.Chan <- MsgSetState{Remove{true}}
                a.svc.Game <- MsgEntityRemoved{&a.player}
                return
            }
        case msg := <-input:
            switch *msg.Type {
            case protocol.Message_Type(protocol.Message_MOVE):
                dir := msg.Move.Direction
                vec := s3dm.NewV3(*dir.X, *dir.Y, *dir.Z)
                a.player.Chan <- MsgRunAction{Move{vec}, false}
            default:
                log.Println("Client sent unhandled message, ignoring:",
                    protocol.Message_Type_name[int32(*msg.Type)])

            }
        }
    }
}
