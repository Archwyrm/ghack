// Copyright 2010 The ghack Authors. All rights reserved.
// Use of this source code is governed by the GNU General Public License
// version 3 (or any later version). See the file COPYING for details.

// Namespace for message identification
package msgId

type MsgId int

const (
    // Component messages
    Tick      = iota // Signal to the component that it should update
    GetState         // Request a State to be returned on StateReply chan
    AddAction        // Add some kind of action to the Entity's list

    // PubSub messages
    Publish     // Publish a message to a topic
    Subscribe   // Subscribe to a topic
    Unsubscribe // Unsubscribe from a topic
)
