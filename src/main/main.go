// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Server main.
package main

import (
    .   "core"
    "game"
    "comm"
    "pubsub"
    "sf"
)

func main() {
    svc := NewServiceContext()

    comm.AvatarFunc = sf.MakeAvatar
    go comm.NewCommService(svc, "0.0.0.0:9190").Run(svc.Comm)
    go pubsub.NewPubSub(svc).Run(svc.PubSub)
    go sf.NewWorld(svc).Run(svc.World)

    game.InitFunc = initGameSvc
    game := game.NewGame(svc)

    game.Run(svc.Game)
}

// Initialize the game with some default data. Eventually this will come from
// data files and those will be loaded elsewhere.
func initGameSvc(g *game.Game, svc ServiceContext) {
    spider := sf.InitSpider(g.GetUid())
    g.AddEntity(spider)
    go spider.Run(svc)
}
