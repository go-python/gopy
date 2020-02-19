# Copyright 2020 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import gopygc
import _gopygc

print(_gopygc.NumHandles())


# test literals
a = gopygc.StructA()
b = gopygc.SliceA()
c = gopygc.MapA()
print(_gopygc.NumHandles())
del a
del b
del c

print(_gopygc.NumHandles())
a = [gopygc.StructValue(), gopygc.StructValue(), gopygc.StructValue()]
print(_gopygc.NumHandles())  # 3
b = [gopygc.SliceScalarValue(), gopygc.SliceScalarValue()]
print(_gopygc.NumHandles())  # 5
c = gopygc.SliceStructValue()
print(_gopygc.NumHandles())  # 6
d = gopygc.MapValue()
print(_gopygc.NumHandles())  # 7
e = gopygc.MapValueStruct()
print(_gopygc.NumHandles())  # 8

del a
print(_gopygc.NumHandles())  # 5
del b
print(_gopygc.NumHandles())  # 3
del c
print(_gopygc.NumHandles())  # 2
del d
print(_gopygc.NumHandles())  # 1
del e
print(_gopygc.NumHandles())  # 0

e1 = gopygc.ExternalType()
print(_gopygc.NumHandles())  # 1
del e1
print(_gopygc.NumHandles())  # 0

# test reference counting
f = gopygc.SliceStructValue()
print(_gopygc.NumHandles())  # 1
g = gopygc.StructA(handle=f.handle)
print(_gopygc.NumHandles())  # 1
del g
print(_gopygc.NumHandles())  # 1
del f
print(_gopygc.NumHandles())  # 0


print("OK")
