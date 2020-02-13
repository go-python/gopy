# Copyright 2020 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import cstrings
import sys
import gc

v = cstrings.StringValue()
print(sys.getrefcount(v))
print(v)
print(gc.is_tracked(v))

del v
print("OK")
