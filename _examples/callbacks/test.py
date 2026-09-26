# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function

import io
import sys

import callbacks

print("--- Each: int and string arguments")
callbacks.Each(3, lambda i, label: print("each:", i, label))

print("--- Mixed: bool, float and uint8 arguments")
# bool() because the pybindgen backend passes a bool as 1 or 0
callbacks.Mixed(lambda on, x, u: print("mixed:", bool(on), x, u))

print("--- Twice: no arguments")
calls = []
callbacks.Twice(lambda: calls.append(1))
print("twice:", len(calls))

print("--- Counter.Visit: a Go struct arrives as a handle")
c = callbacks.Counter()

def visit(handle, n):
    seen = callbacks.Counter(handle=handle)
    print("visit:", n, seen.N)

c.Visit(2, visit)
print("counter:", c.N)

print("--- Describe: an interface{} arrives as a string")
callbacks.Describe(lambda v: print("describe:", repr(v)))

print("--- Count: a bool result")
print("count:", callbacks.Count(10, lambda i: i % 3 == 0))

print("--- Sum: an int result")
print("sum:", callbacks.Sum(5, lambda i: i * i))

print("--- Widest: a uint result")
print("widest:", callbacks.Widest(4, lambda i: i * 10))

print("--- Apply: a float result")
print("apply:", callbacks.Apply(1.5, lambda x: x * 2))

print("--- Counter.Check: a handle argument and a bool result")
print("check:", c.Check(lambda handle, n: callbacks.Counter(handle=handle).N == n))

print("--- a bound method")

class Box(object):
    def __init__(self):
        self.items = []

    def add(self, i, label):
        self.items.append((i, label))

box = Box()
callbacks.Each(2, box.add)
print("box:", box.items)

print("--- called from another goroutine")
callbacks.InGoroutine(lambda i: print("goroutine:", i))

print("--- an exception in a callback is reported, and Go carries on")
seen = []

def boom(i, label):
    seen.append(i)
    raise ValueError("boom %d" % i)

stderr, sys.stderr = sys.stderr, io.StringIO()
try:
    callbacks.Each(3, boom)
    reported = sys.stderr.getvalue()
finally:
    sys.stderr = stderr
print("calls:", len(seen), "reported:", reported.count("ValueError: boom"))

print("OK")
