# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import consts

print("c1 = %s" % consts.C1)
print("c2 = %s" % consts.C2)
print("c3 = %s" % consts.C3)
print("c4 = %s" % consts.C4)
print("c5 = %s" % consts.C5)
print("c6 = %s" % consts.C6)
print("c7 = %s" % consts.C7)

print("k1 = %s" % consts.Kind1)
print("k2 = %s" % consts.Kind2)
## FIXME: unexported types not supported yet (issue #44)
#print("k3 = %s" % consts.Kind3)
#print("k4 = %s" % consts.Kind4)

print("OK")
