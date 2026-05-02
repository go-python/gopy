# Copyright 2026 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

# Regression test for GIL ordering bug (issue #370).
# Exact reproduction from the issue report: two separately-built gopy
# extensions loaded in the same process, with calls interleaved in a loop.
from gilstring.gilstring import Hello
from simple.simple import Add

for _ in range(5000):
    Add(2, 2)
    Hello('hi')

print("OK")
