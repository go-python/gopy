# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Regression/stress test for the complex64/complex128 GIL-handling fix.
#
# Comp64Add/Comp128Add marshal a raw *C.PyObject on both sides of the call:
# PyComplex_AsCComplex on the way in, PyComplex_FromDoubles on the way out.
# Both must run while the GIL is held -- gopy releases the GIL only around
# the pure-Go call in between. This hammers the round trip from the main
# thread while a background thread continuously allocates and collects
# Python objects, so that if the marshalling window ever slipped outside the
# GIL-held region, refcount/GC corruption would have a chance to surface.

import gc
import sys
import threading

import simple as pkg

ITERATIONS = 500
STOP = threading.Event()


def churn():
    while not STOP.is_set():
        _ = [object() for _ in range(100)]
        gc.collect(0)


t = threading.Thread(target=churn, daemon=True)
t.start()

try:
    for i in range(ITERATIONS):
        a = complex(i, -i)
        b = complex(-i, i * 2)
        want = a + b

        got64 = pkg.Comp64Add(a, b)
        # complex64 loses precision relative to Python's complex128
        # arithmetic; compare with a tolerance.
        if abs(got64 - want) > 1e-3:
            print("FAIL: Comp64Add(%s, %s) = %s, want ~%s" % (a, b, got64, want),
                  file=sys.stderr)
            sys.exit(1)

        got128 = pkg.Comp128Add(a, b)
        if got128 != want:
            print("FAIL: Comp128Add(%s, %s) = %s, want %s" % (a, b, got128, want),
                  file=sys.stderr)
            sys.exit(1)

        if i % 50 == 0:
            gc.collect()
finally:
    STOP.set()
    t.join()

print("OK")
