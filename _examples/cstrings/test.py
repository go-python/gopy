# Copyright 2020 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import cstrings
import gc
import resource

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
    m = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
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
    a = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    pass1 = _run_fn(fn)
    b = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    pass2 = _run_fn(fn)
    c = resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    if verbose:
        print(fn.__name__, pass1)
        print(fn.__name__, pass2)
        print(fn.__name__, a, b, c)

    print(fn.__name__,  "leaked: ", (c-b) > (size * iterations))

    # bump up the size of each successive test to ensure that leaks
    # are not absorbed by previous rss growth.
    size += 4096


print("OK")
