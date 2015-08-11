# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import structs

print("s = structs.S()")
s = structs.S()
print("s = %s" % (s,))
print("s.Init()")
s.Init()
print("s.Upper('boo')= %r" % (s.Upper("boo"),))

