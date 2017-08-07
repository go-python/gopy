# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import vars

print("doc(vars):\n%s" % repr(vars.__doc__))
print("doc(vars.GetV1()):\n%s" % repr(vars.GetV1.__doc__))
print("doc(vars.SetV1()):\n%s" % repr(vars.SetV1.__doc__))

print("Initial values")
print("v1 = %s" % vars.GetV1())
print("v2 = %s" % vars.GetV2())
print("v3 = %s" % vars.GetV3())
print("v4 = %s" % vars.GetV4())
print("v5 = %s" % vars.GetV5())
print("v6 = %s" % vars.GetV6())
print("v7 = %s" % vars.GetV7())

print("k1 = %s" % vars.GetKind1())
print("k2 = %s" % vars.GetKind2())
## FIXME: unexported types not supported yet (issue #44)
#print("k3 = %s" % vars.GetKind3())
#print("k4 = %s" % vars.GetKind4())

vars.SetV1("test1")
vars.SetV2(90)
vars.SetV3(1111.1111)
vars.SetV4("test2")
vars.SetV5(50)
vars.SetV6(50)
vars.SetV7(1111.1111)

vars.SetKind1(123)
vars.SetKind2(456)
print("New values")
print("v1 = %s" % vars.GetV1())
print("v2 = %s" % vars.GetV2())
print("v3 = %s" % vars.GetV3())
print("v4 = %s" % vars.GetV4())
print("v5 = %s" % vars.GetV5())
print("v6 = %s" % vars.GetV6())
print("v7 = %s" % vars.GetV7())

print("k1 = %s" % vars.GetKind1())
print("k2 = %s" % vars.GetKind2())

print("vars.GetDoc() = %s" % repr(vars.GetDoc()))
print("doc of vars.GetDoc = %s" % (repr(vars.GetDoc.__doc__),))
print("doc of vars.SetDoc = %s" % (repr(vars.SetDoc.__doc__),))
