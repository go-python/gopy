# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import cgo

print("cgo.doc: %r" % (cgo.__doc__,))
print("cgo.Hi()= %r" % (cgo.Hi(),))
print("cgo.Hello(you)= %r" % (cgo.Hello("you"),))
