# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import gostrings

print("S1 = %s" % (gostrings.S1(),))
print("GetString() = %s" % (gostrings.GetString(),))

print("OK")
