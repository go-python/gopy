# Copyright 2015-present, The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

import packet

p = packet.NewPing(7)
print("GoPacketToBytes(3, p) = %s" % bytes(packet.GoPacketToBytes(3, p)).decode())
print("GoPacketToBytes(shieldID=4, packet=p) = %s" % bytes(packet.GoPacketToBytes(shieldID=4, packet=p)).decode())
print("PingID(p) = %d" % packet.PingID(p))
print("Double(21) = %d" % packet.Double(21))
print("p.Add(3) = %d" % p.Add(3))
print("Double(packet=5) = %d" % packet.Double(packet=5))
print("p.Add(packet=1) = %d" % p.Add(packet=1))
print("Apply(4, square) = %d" % packet.Apply(4, lambda i: i * i))
