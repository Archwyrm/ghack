// Copyright 2010, 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package core

type Service interface {
    Run(input chan Msg)
}

type ServiceContext struct {
    Game, Comm, PubSub chan Msg
}

func NewServiceContext() ServiceContext {
    return ServiceContext{make(chan Msg), make(chan Msg), make(chan Msg)}
}
