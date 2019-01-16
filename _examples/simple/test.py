# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import simple as pkg

print("doc(pkg):\n%s" % repr(pkg.__doc__).lstrip('u'))
print("pkg.Func()...")
pkg.Func()
print("fct = pkg.Func...")
fct = pkg.Func
print("fct()...")
fct()

print("pkg.Add(1,2)= %s" % (pkg.Add(1,2),))
print("pkg.Bool(True)= %s" % pkg.Bool(True))
print("pkg.Bool(False)= %s" % pkg.Bool(False))
a = 3+4j
b = 2+5j
print("pkg.Comp64Add(%s, %s) = %s" % (a, b, pkg.Comp128Add(a, b)))
print("pkg.Comp128Add(%s, %s) = %s" % (a, b, pkg.Comp128Add(a, b)))

print("OK")
