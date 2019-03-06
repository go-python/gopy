# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function
import sys

_PY3 = sys.version_info[0] == 3
if _PY3:
    xrange = range
    pass

import named

### test docs
print("doc(named): %s" % repr(named.__doc__).lstrip('u'))
# print("doc(named.Float): %s" % repr(named.Float.__doc__).lstrip('u'))
# print("doc(named.Float.Value): %s" % repr(named.Float.Value.__doc__).lstrip('u'))
# 
# print("v = named.Float()")
# v = named.Float()
# print("v = %s" % (v,))
# print("v.Value() = %s" % (v.Value(),))
# 
# print("x = named.X()")
# x = named.X()
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XX()")
# x = named.XX()
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XXX()")
# x = named.XXX()
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XXXX()")
# x = named.XXXX()
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# ### test ctors
# 
# print("v = named.Float(42)")
# v = named.Float(42)
# print("v = %s" % (v,))
# print("v.Value() = %s" % (v.Value(),))
# 
# print("v = named.Float(42.0)")
# v = named.Float(42.0)
# print("v = %s" % (v,))
# print("v.Value() = %s" % (v.Value(),))
# 
# print("x = named.X(42)")
# x = named.X(42)
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XX(42)")
# x = named.XX(42)
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XXX(42)")
# x = named.XXX(42)
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XXXX(42)")
# x = named.XXXX(42)
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("x = named.XXXX(42.0)")
# x = named.XXXX(42.0)
# print("x = %s" % (x,))
# print("x.Value() = %s" % (x.Value(),))
# 
# print("s = named.Str()")
# s = named.Str()
# print("s = %s" % (s,))
# print("s.Value() = %s" % repr(s.Value()).lstrip('u'))
# 
# print("s = named.Str('string')")
# s = named.Str("string")
# print("s = %s" % (s,))
# print("s.Value() = %s" % repr(s.Value()).lstrip('u'))

# note: cannot construct arrays from python -- too risky wrt len etc -- use slices

# print("arr = named.Array()")
# arr = named.Array()
# print("arr = %s" % (arr,))

# print("arr = named.Array([1,2])")
# arr = named.Array([1,2])
# print("arr = %s" % (arr,))
# 
# try:
#     print("arr = named.Array(range(10))")
#     arr = named.Array(range(10))
#     print("arr = %s" % (arr,))
# except Exception as err:
#     print("caught: %s" % (str(err),))
#     pass

# print("arr = named.Array(xrange(2))")
# arr = named.Array(xrange(2))
# print("arr = %s" % (arr,))
# 
print("s = named.Slice()")
s = named.Slice()
print("s = %s" % (s,))

print("s = named.Slice([1,2])")
s = named.Slice([1,2])
print("s = %s" % (s,))

print("s = named.Slice(range(10))")
s = named.Slice(range(10))
print("s = %s" % (s,))

print("s = named.Slice(xrange(10))")
s = named.Slice(xrange(10))
print("s = %s" % (s,))

print("OK")
