# Copyright 2015-present, The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Regression test for the GIL handling of a Python callback's return value.
#
# When Go calls a Python callback, the generated wrapper takes the GIL, calls
# the callback, and converts the returned object to a Go value. That
# conversion, and the decref of the returned object, must both happen before
# the wrapper releases the GIL again.

import contextlib
import io
import os
import sys

import callbacks


def quietly(f, *args):
    """Calls f(*args), and returns its result with what it wrote to stderr,
    where a callback's error gets printed."""
    err = io.StringIO()
    with contextlib.redirect_stderr(err):
        got = f(*args)
    return got, err.getvalue()


def boom(i):
    raise ValueError("boom")


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

# A callback with no result must not leak a reference to what it returns.
L = []
before = sys.getrefcount(L)
for _ in range(100):
    callbacks.Twice(lambda: L)
delta = sys.getrefcount(L) - before
if delta != 0:
    raise AssertionError("refcount of the returned list grew by %d over 200 calls" % delta)

# The pybind11 and nanobind trampolines abort the process when a callback
# raises, or returns something that won't convert.
if os.environ.get("GOPY_BACKEND") not in ("pybind11", "nanobind"):
    # A result that fails to convert gets reported like a callback's own
    # error, not left pending for the next Python code to trip over.
    _, err = quietly(callbacks.Apply, 1.5, lambda x: "abc")
    if "TypeError" not in err:
        raise AssertionError("Apply(1.5, lambda x: 'abc') reported %r, want a TypeError" % err)

    # A callback that raises has no result to convert: Go gets the zero
    # value.
    got, err = quietly(callbacks.Sum, 3, boom)
    if got != 0 or "ValueError: boom" not in err:
        raise AssertionError("Sum(3, boom) = %r, reporting %r; want 0, reporting the ValueError" % (got, err))

    # Converting a missing string result segfaulted.
    if hasattr(callbacks, "Label"):
        got, err = quietly(callbacks.Label, 1, boom)
        if got != "" or "ValueError: boom" not in err:
            raise AssertionError("Label(1, boom) = %r, reporting %r; want '', reporting the ValueError" % (got, err))

print("OK")
