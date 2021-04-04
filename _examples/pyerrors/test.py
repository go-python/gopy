# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import pyerrors

def div(a, b):
    try:
        r = pyerrors.Div(a, b)
        print("pyerrors.Div(%d, %d) = %d"% (a, b, r))
    except Exception as e:
        print(e)


def new_mystring(s):
    try:
        ms = pyerrors.NewMyString(s)
        print('pyerrors.NewMyString("%s") = "%s"'% (s, ms.String()))
    except Exception as e:
        print(e)


div(5,0)  # error
div(5,2)

new_mystring("")  # error
new_mystring("hello")

print("OK")
