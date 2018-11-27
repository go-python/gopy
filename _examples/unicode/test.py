# -*- coding: utf-8 -*-

# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function, unicode_literals

import encoding

bytestr = b"Python byte string"
unicodestr = u"Python Unicode string 🐱"

bytestr_ret = encoding.HandleString(bytestr)
unicodestr_ret = encoding.HandleString(unicodestr)

print("encoding.HandleString(bytestr) ->", bytestr_ret)
print("encoding.HandleString(unicodestr) ->", unicodestr_ret)

gostring_ret = encoding.GetString()
print("encoding.GetString() ->", gostring_ret)

