# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import iface, go

### test docs
print("doc(iface): %r" % (iface.__doc__,))

print("t = iface.T()")
t = iface.T()
print("t.F()")
t.F()

print("iface.CallIface(t)")
iface.CallIface(t)

print('iface.IfaceString("test string"')
iface.IfaceString("test string")

print('iface.IfaceString(str(42))')
iface.IfaceString(str(42))

print('iface.IfaceHandle(t)')
iface.IfaceHandle(t)

print('iface.IfaceHandle(go.nil)')
iface.IfaceHandle(go.nil)

print("OK")
