# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import gilstring

# Regression test for GIL ordering bug (issue #370 / PR #386).
# Mirrors the exact reproduction script from the issue report:
#   for _ in range(5000):
#       Add(2, 2)
#       Hello('hi')
# Integer arithmetic (Add) interleaved with string-returning calls (Hello)
# stresses the go2py C.CString path that previously ran without the GIL held.
N = 5000

for _ in range(N):
    assert gilstring.Add(2, 2) == 4
    assert gilstring.Hello("hi") == "Hello, hi!"

# Struct string fields, slice elements, and map values exercise additional
# go2py string conversion paths from the same bug.
for _ in range(N):
    item = gilstring.MakeItem("hello")
    assert item.Label == "hello", item.Label

for _ in range(N):
    s = gilstring.StringSlice()
    assert s[0] == "alpha", s[0]

for _ in range(N):
    m = gilstring.StringMap()
    assert m["key"] == "value", m["key"]

print("OK")
