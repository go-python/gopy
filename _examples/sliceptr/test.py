# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import sliceptr

i = sliceptr.IntVector()

sliceptr.Fill(i)
print(i)
sliceptr.Append(i)
print(i)
s = sliceptr.StrVector()
sliceptr.Convert(i, s)
print(s)

print("OK")
