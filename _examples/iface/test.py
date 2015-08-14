# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import iface

### test docs
print("doc(iface): %r" % (iface.__doc__,))

print("t = iface.T()")
t = iface.T()
print("t.F()")
t.F()

print("iface.CallIface(t)")
iface.CallIface(t)

