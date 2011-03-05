ghack server
===============================================================================
git version - http://dungeonhack.sourceforge.net


About
-------------------------------------------------------------------------------
ghack is the first server component of project DungeonHack's new
architecture. It is intended to be highly flexible, robust, network
oriented, and concurrent. DungeonHack is a first person, action oriented
role-playing game and ghack focuses on supporting this, however it is
possible to create variations of this or other kinds of games. Game code
is built using a component aggregation framework and adding new game
logic is achieved by simply adding new components. Extending game logic
through components will also be possible using a scripting language of
choice (plans are to start with Python).

Current code is in a very skeletal state, however.


License
-------------------------------------------------------------------------------
Copyright 2010, 2011 The ghack Authors. All rights reserved.

Unless otherwise noted, source code and assets are licensed under
the terms of the GNU General Public License as published by the Free
Software Foundation; either version 3 of the License, or (at your option)
any later version.

Please see the file COPYING for more details.


Building and Running
-------------------------------------------------------------------------------
This software is known to work on GNU/Linux, probably works on Mac OS X,
and *might* work on Windows. Future versions will support all these
operating systems.

In short, the following things are required to use this software:

 * [Go compiler](http://golang.org)
 * [godag build tool](http://code.google.com/p/godag/)
 * [protobuf - Protocol Buffers](http://code.google.com/p/protobuf/)
 * [goprotobuf](http://code.google.com/p/goprotobuf/)
 * [s3dm - Simple 3D Maths](https://github.com/tm1rbrt/s3dm)

Please see the individual package's installation instructions in order
to install. All Go dependencies should use the latest version available.

Some dependencies may be installed easily with goinstall:

    goinstall goprotobuf.googlecode.com/hg/proto
    goinstall github.com/tm1rbrt/s3dm

Once all the dependencies are present, run:

    protoc --go_out=src/ protocol/protocol.proto
    gd -o ghack

If successful, the output binary 'ghack' may be run.

