# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
import ifxmap

im = ifxmap.New()

assert(isinstance(im["float"], float))
assert(im["float"] == 3.14159)

assert(isinstance(im["string"], str))
assert(im["string"] == "sample")

assert(isinstance(im["int"], int))
assert(im["int"] == 2)

assert(isinstance(im["bool"], bool))
assert(im["bool"] == True)

assert(isinstance(im["struct"], ifxmap.TestStruct))
assert(im["struct"].Value == 6)

assert(isinstance(im["structPtr"], ifxmap.TestStruct))
assert(im["structPtr"].Value == 12)

assert(len(im) == 6)

print("OK")
