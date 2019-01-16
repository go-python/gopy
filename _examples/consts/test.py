# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import consts

print("c1 = %s" % consts.GetC1())
print("c2 = %s" % consts.GetC2())
print("c3 = %s" % consts.GetC3())
print("c4 = %s" % consts.GetC4())
print("c5 = %s" % consts.GetC5())
print("c6 = %s" % consts.GetC6())
print("c7 = %s" % consts.GetC7())

print("k1 = %s" % consts.GetKind1())
print("k2 = %s" % consts.GetKind2())
## FIXME: unexported types not supported yet (issue #44)
#print("k3 = %s" % consts.GetKind3())
#print("k4 = %s" % consts.GetKind4())

print("OK")
