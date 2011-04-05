// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Server main.
package main

import (
    "github.com/tm1rbrt/s3dm"
    "core"
    "comm"
    "pubsub"
    "sf"
)

func main() {
    svc := core.NewServiceContext()

    comm.AvatarFunc = sf.MakeAvatar
    go comm.NewCommService(svc, "0.0.0.0:9190").Run(svc.Comm)
    go pubsub.NewPubSub().Run(svc.PubSub)

    game := core.NewGame(svc)
    initGameSvc(game)
    game.Run(svc.Game)
}

// Initialize the game with some default data. Eventually this will come from
// data files and those will be loaded elsewhere.
func initGameSvc(g *core.Game) {
    g.PlayerFunc = playerWrapper // Register the player spawning func
    spider := sf.NewSpider(g.GetUid())
    spiderChan := make(chan core.Msg)
    g.AddEntity(spider, spiderChan)
    inc := 1.0 / 60 // Move by 1 unit/s at rate of 60 ticks/s
    spider.AddAction(sf.Move{s3dm.NewV3(inc, inc, 0)})
    go spider.Run(spiderChan)
}

// Wrap sf.NewPlayer since it returns *sf.Player and not core.Entity
func playerWrapper(uid core.UniqueId) core.Entity {
    return sf.NewPlayer(uid)
}
