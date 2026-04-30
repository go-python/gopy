# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import gilstring

# Regression test for GIL ordering bug (issue #370 / PR #386).
# Struct string fields, slice elements, and map values all previously used
# C.CString (go2py) without the GIL held, causing crashes under repeated calls.
N = 5000

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
