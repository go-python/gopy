# Copyright 2020 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import cstrings
import gc
import sys

# resource module is Unix-only, not available on Windows
# On Windows, use psutil for memory tracking
if sys.platform == 'win32':
    import psutil
    HAS_RESOURCE = False
else:
    import resource
    HAS_RESOURCE = True

verbose = False
iterations = 10000
size = 4096


def gofnString():
    return cstrings.StringValue("a", size)


def gofnStruct():
    s = cstrings.StringInStruct("a", size)
    return s.V


def gofnNestedStruct():
    s = cstrings.StringInNestedStruct("a", size)
    return s.S.V


def gofnSlice():
    s = cstrings.StringSlice("a", size)
    return s[0]


def gofnMap():
    m = cstrings.StringMap("a", size)
    return m["a"]


def print_memory(s):
    if HAS_RESOURCE:
        m = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    else:
        # psutil returns memory in bytes, convert to KB to match resource module
        m = psutil.Process().memory_info().rss // 1024
    if verbose:
        print(s, m)
    return m


def _run_fn(fn):
    memoryvals = []
    t = [fn() for _ in range(iterations)]
    memoryvals.append(print_memory(
        "Memory usage after first list creation is:"))

    t = [fn() for _ in range(iterations)]
    memoryvals.append(print_memory(
        "Memory usage after second list creation is:"))

    gc.collect()
    memoryvals.append(print_memory("Memory usage after GC:"))

    t = [fn() for _ in range(iterations)]
    memoryvals.append(print_memory(
        "Memory usage after third list creation is:"))

    gc.collect()
    memoryvals.append(print_memory("Memory usage after GC:"))
    return memoryvals


for fn in [gofnString, gofnStruct, gofnNestedStruct,  gofnSlice, gofnMap]:
    alloced = size * iterations
    a = print_memory("Initial memory:")
    pass1 = _run_fn(fn)
    b = print_memory("After first pass:")
    pass2 = _run_fn(fn)
    c = print_memory("After second pass:")
    if verbose:
        print(fn.__name__, pass1)
        print(fn.__name__, pass2)
        print(fn.__name__, a, b, c)

    leaked = (c-b) > (size * iterations)
    print(fn.__name__,  "leaked: ", leaked)

    # bump up the size of each successive test to ensure that leaks
    # are not absorbed by previous rss growth.
    size += 4096


print("OK")
