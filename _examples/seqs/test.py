# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import sys
_PY3 = sys.version_info[0] == 3
if _PY3:
    xrange = range

import seqs

### test docs
print("doc(seqs): %s" % repr(seqs.__doc__).lstrip('u'))

# note: arrays not settable from python -- use slices instead

# print("arr = seqs.Array(xrange(2))")
# arr = seqs.Array(xrange(2))
# print("arr = %s" % (arr,))
# 
print("s = seqs.Slice()")
s = seqs.Slice()
print("s = %s" % (s,))

print("s = seqs.Slice([1,2])")
s = seqs.Slice([1,2])
print("s = %s" % (s,))

print("s = seqs.Slice(range(10))")
s = seqs.Slice(range(10))
print("s = %s" % (s,))

print("s = seqs.Slice(xrange(10))")
s = seqs.Slice(xrange(10))
print("s = %s" % (s,))

print("s = seqs.Slice()")
s = seqs.Slice()
print("s = %s" % (s,))
print("s += [1,2]")
s += [1,2]
print("s = %s" % (s,))
print("s += [10,20]")
s += [10,20]
print("s = %s" % (s,))

print("OK")
