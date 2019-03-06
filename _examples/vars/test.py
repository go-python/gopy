# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import vars

print("doc(vars):\n%s" % repr(vars.__doc__).lstrip('u'))
print("doc(vars.V1()):\n%s" % repr(vars.V1.__doc__).lstrip('u'))
print("doc(vars.Set_V1()):\n%s" % repr(vars.Set_V1.__doc__).lstrip('u'))

print("Initial values")
print("v1 = %s" % vars.V1())
print("v2 = %s" % vars.V2())
print("v3 = %s" % vars.V3())
print("v4 = %s" % vars.V4())
print("v5 = %s" % vars.V5())
print("v6 = %s" % vars.V6())
print("v7 = %s" % vars.V7())

print("k1 = %s" % vars.Kind1())
print("k2 = %s" % vars.Kind2())
## FIXME: unexported types not supported yet (issue #44)
#print("k3 = %s" % vars.Kind3())
#print("k4 = %s" % vars.Kind4())

vars.Set_V1("test1")
vars.Set_V2(90)
vars.Set_V3(1111.1111)
vars.Set_V4("test2")
vars.Set_V5(50)
vars.Set_V6(50)
vars.Set_V7(1111.1111)

vars.Set_Kind1(123)
vars.Set_Kind2(456)
print("New values")
print("v1 = %s" % vars.V1())
print("v2 = %s" % vars.V2())
print("v3 = %s" % vars.V3())
print("v4 = %s" % vars.V4())
print("v5 = %s" % vars.V5())
print("v6 = %s" % vars.V6())
print("v7 = %s" % vars.V7())

print("k1 = %s" % vars.Kind1())
print("k2 = %s" % vars.Kind2())

print("vars.Doc() = %s" % repr(vars.Doc()).lstrip('u'))
print("doc of vars.Doc = %s" % repr(vars.Doc.__doc__).lstrip('u'))
print("doc of vars.Set_Doc = %s" % repr(vars.Set_Doc.__doc__).lstrip('u'))

print("OK")
