# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import pointers

print("s = pointers.S(2)")
s = pointers.S(2)
print("s = %s" % (s,))
print("s.Value = %s" % (s.Value,))

print("pointers.Inc(s)")
print("s.Value = %s" % (s.Value,))

print("OK")
