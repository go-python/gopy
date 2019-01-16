# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import cgo

print("cgo.doc: %s" % repr(cgo.__doc__).lstrip('u'))
print("cgo.Hi()= %s" % repr(cgo.Hi()).lstrip('u'))
print("cgo.Hello(you)= %s" % repr(cgo.Hello("you")).lstrip('u'))

print("OK")
