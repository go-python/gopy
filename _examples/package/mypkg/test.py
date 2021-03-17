# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import absolute_import, print_function

import mypkg.mypkg

print("mypkg.mypkg.SayHello()...")
print(mypkg.mypkg.SayHello())
print("OK")
