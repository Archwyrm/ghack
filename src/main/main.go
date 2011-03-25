// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Server main.
package main

import (
    "core"
    "comm"
    "pubsub"
)

func main() {
    svc := core.NewServiceContext()

    go comm.NewCommService(svc, ":9190").Run(svc.Comm)
    go pubsub.NewPubSub().Run(svc.PubSub)

    game := NewGame(svc)
    game.Run(svc.Game)
}
