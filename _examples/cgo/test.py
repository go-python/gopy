# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import cgo

print("cgo.doc: %s" % repr(cgo.__doc__).lstrip('u'))
print("cgo.Hi()= %s" % repr(cgo.Hi()).lstrip('u'))
print("cgo.Hello(you)= %s" % repr(cgo.Hello("you")).lstrip('u'))

# Regression test for GIL ordering bug (issue #370 / PR #386):
# go functions returning string (go2py=C.CString) previously called C.CString
# without holding the GIL, causing crashes under repeated calls.
for _ in range(5000):
    assert cgo.Hi() == 'hi from go\n'
    assert cgo.Hello("world") == 'hello world from go\n'
print("stress OK")

print("OK")
