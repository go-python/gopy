# Copyright 2015-present, The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Regression test for the GIL handling of a Python callback's return value.
#
# When Go calls a Python callback, the generated wrapper takes the GIL, calls
# the callback, and converts the returned object to a Go value. That
# conversion, and the decref of the returned object, must both happen before
# the wrapper releases the GIL again.

import sys

import callbacks

# A callback that keeps returning the same float must not leak a reference to
# it on each call.
X = float("0.25")
before = sys.getrefcount(X)
for _ in range(100):
    callbacks.Apply(1.5, lambda x: X)
delta = sys.getrefcount(X) - before
if delta != 0:
    raise AssertionError("refcount of the returned float grew by %d over 100 calls" % delta)

# A float64 callback that returns an int makes the conversion
# (PyFloat_AsDouble) run int.__float__, which allocates a Python float. That
# segfaults if the GIL has already been released.
got = callbacks.Apply(1.5, lambda x: 3)
if got != 3.0:
    raise AssertionError("Apply(1.5, lambda x: 3) = %r, want 3.0" % got)

print("OK")
