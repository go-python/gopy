# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import gopygc
import _gopygc

print(_gopygc.NumHandles())
a = [gopygc.StructValue(), gopygc.StructValue(), gopygc.StructValue()]
print(_gopygc.NumHandles())
b = [gopygc.SliceScalarValue(), gopygc.SliceScalarValue()]
print(_gopygc.NumHandles())
c = gopygc.SliceStructValue()
print(_gopygc.NumHandles())
d = gopygc.MapValue()
print(_gopygc.NumHandles())
e = gopygc.MapValueStruct()
print(_gopygc.NumHandles())

del a
print(_gopygc.NumHandles())
del b
print(_gopygc.NumHandles())
del c
print(_gopygc.NumHandles())
del d
print(_gopygc.NumHandles())
del e
print(_gopygc.NumHandles())

print("OK")
