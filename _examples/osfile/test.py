# Copyright 2019 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

# note: this doesn't work without actually doing "make install" in pkg build dir
# and furthermore, os is also used in python, so probably not very practical
# to include it in the first place -- need to use a wrapper solution.

from osfile import osfile, os, go

try:
    f = os.Create("testfile.test")
except Exception as err:
    print("os.Create got an error: %s" % (err,))
    pass

    
try:
    f.Write("this is a test of python writing to a Go-opened file\n")
except Exception as err:
    print("file.Write got an error: %s" % (err,))
    pass

f.Close()

try:
    os.Remove("testfile.test")
except Exception as err:
    print("os.Remove got an error: %s" % (err,))
    pass


