// Copyright 2011 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

package core

import "core/cmpId"

// Signifies that the owner entity is in the process of being removed.
type Remove struct {
    Remove bool
}

func (s Remove) Id() StateId  { return cmpId.Remove }
func (s Remove) Name() string { return "Remove" }
