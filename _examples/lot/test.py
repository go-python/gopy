# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
import lot

l = lot.New()
print('l.SomeString : {}'.format(l.SomeString))
print('l.SomeInt : {}'.format(l.SomeInt))
print('l.SomeFloat : {}'.format(l.SomeFloat))
print('l.SomeBool : {}'.format(l.SomeBool))
print('l.SomeListOfStrings: {}'.format(l.SomeListOfStrings))
print('l.SomeListOfInts: {}'.format(l.SomeListOfInts))
print('l.SomeListOfFloats: {}'.format(l.SomeListOfFloats))
print('l.SomeListOfBools: {}'.format(l.SomeListOfBools))

print("OK")
