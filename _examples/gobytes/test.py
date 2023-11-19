# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
import gobytes, go

a = bytes([0, 1, 2, 3])
b = gobytes.CreateBytes(10)
print ("Python bytes:", a)
print ("Go slice: ", b)

print ("gobytes.HashBytes from Go bytes:", gobytes.HashBytes(b))

print("Python bytes to Go: ", go.Slice_byte.from_bytes(a))
print("Go bytes to Python: ", bytes(go.Slice_byte([3, 4, 5])))

print("OK")
