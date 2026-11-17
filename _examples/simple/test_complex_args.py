# Copyright 2015-present, The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Regression test for the complex64/complex128 GIL-handling fix.
#
# A complex argument is converted with PyComplex_AsCComplex. For an exact
# `complex` that is a plain read of two doubles and works even without the
# GIL, so passing complex values cannot catch the bug. For a float or an int
# it runs Python code (the __complex__ lookup, then __float__/__index__),
# which segfaults if the GIL has already been released for the wrapped Go
# call. Each call below crashed before the fix and returns the right value
# after it.

import simple as pkg

cases = [
    ("float", 1.5, 2.5, 4 + 0j),
    ("int", 1, 2, 3 + 0j),
]

for name, a, b, want in cases:
    for fn in (pkg.Comp64Add, pkg.Comp128Add):
        got = fn(a, b)
        # every expected value is exactly representable as a complex64
        if got != want:
            raise AssertionError("%s(%s args) = %s, want %s" % (fn.__name__, name, got, want))

print("OK")
