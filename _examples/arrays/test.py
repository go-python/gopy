# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
import arrays

a = [1,2,3,4]
b = arrays.CreateArray()
print ("Python list:", a)
print ("Go array: ", b)
# note: only support slices as converted from python -- arrays don't make sense
#print ("arrays.IntSum from Python list:", arrays.IntSum(a))
print ("arrays.IntSum from Go array:", arrays.IntSum(b))

print("OK")
