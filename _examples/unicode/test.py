# -*- coding: utf-8 -*-

# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function, unicode_literals

import sys

import encoding

# There is no portable way of outputting Unicode in Python 2/3 -- we encode
# them manually and sidestep the builtin encoding machinery
try:
    binary_stdout = sys.stdout.buffer  # Python 3
except AttributeError:
    binary_stdout = sys.stdout  # Python 2

bytestr = b"Python byte string"
unicodestr = u"Python Unicode string 🐱"

# TODO: need conversion from bytestr to string -- not sure pybindgen can do it?
#bytestr_ret = encoding.HandleString(bytestr)
unicodestr_ret = encoding.HandleString(unicodestr)

# binary_stdout.write(b"encoding.HandleString(bytestr) -> ")
# binary_stdout.write(bytestr_ret.encode('UTF-8'))
# binary_stdout.write(b'\n')
binary_stdout.write(b"encoding.HandleString(unicodestr) -> ")
binary_stdout.write(unicodestr_ret.encode('UTF-8'))
binary_stdout.write(b'\n')

gostring_ret = encoding.GetString()
binary_stdout.write(b"encoding.GetString() -> ")
binary_stdout.write(gostring_ret.encode("UTF-8"))
binary_stdout.write(b"\n")

print("OK")
