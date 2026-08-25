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

# Regression test for issue #359: 0-length slice must not crash
empty = gobytes.CreateBytes(0)
print("Go empty slice: ", empty)
empty_bytes = bytes(empty)
assert empty_bytes == b"", "expected b'', got %r" % (empty_bytes,)
print("Go empty bytes to Python: ", empty_bytes)

# Regression test for issue #359 (reverse direction): converting a 0-length
# Python bytes to a Go []byte, then back, must not crash either.
empty_from_bytes = go.Slice_byte.from_bytes(b"")
print("Python empty bytes to Go: ", empty_from_bytes)
roundtrip_bytes = bytes(empty_from_bytes)
assert roundtrip_bytes == b"", "expected b'', got %r" % (roundtrip_bytes,)
print("Go empty bytes round-trip: ", roundtrip_bytes)

print("OK")
