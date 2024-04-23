# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
import math
import random
import slices, go

a = [1,2,3,4]
b = slices.CreateSlice()
print ("Python list:", a)
print ("Go slice: ", b)
print ("slices.IntSum from Python list:", slices.IntSum(go.Slice_int(a)))
print ("slices.IntSum from Go slice:", slices.IntSum(b))

su8 = slices.SliceUint8([1,2])
su16 = slices.SliceUint16([2,3])
su32 = slices.SliceUint32([3,4])
su64 = slices.SliceUint64([4,5])
print ("unsigned slice elements:", su8[0], su16[0], su32[0], su64[0])

si8 = slices.SliceInt8([-1,-2])
si16 = slices.SliceInt16([-2,-3])
si32 = slices.SliceInt32([-3,-4])
si64 = slices.SliceInt64([-4,-5])
print ("signed slice elements:", si8[0], si16[0], si32[0], si64[0])

ss = slices.CreateSSlice()
print ("struct slice: ", ss)
print ("struct slice[0]: ", ss[0])
print ("struct slice[1]: ", ss[1])
print ("struct slice[2].Name: ", ss[2].Name)

slices.PrintSSlice(ss)

slices.PrintS(ss[0])
slices.PrintS(ss[1])

cmplx = slices.SliceComplex([(random.random() + random.random() * 1j) for _ in range(16)])
sqrts = slices.CmplxSqrt(cmplx)
for root, orig in zip(sqrts, cmplx):
    root_squared = root * root
    assert math.isclose(root_squared.real, orig.real)
    assert math.isclose(root_squared.imag, orig.imag)


matrix = slices.GetEmptyMatrix(4,4)
for i in range(4):
    for j in range(4):
        assert not matrix[i][j]
print("[][]bool working as expected")

print("OK")
