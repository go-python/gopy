# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import vars

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

vars.SetV1("-v1-")
vars.SetV2(4242)
vars.SetV3(-vars.GetV3())
vars.SetV4("-c4-")
vars.SetV5(24)
vars.SetV6(24)
vars.SetV7(-vars.GetV7())

vars.SetKind1(11)
vars.SetKind2(22)

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


