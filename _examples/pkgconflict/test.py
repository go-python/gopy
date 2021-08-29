# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import pkgconflict

txt = pkgconflict.TextTempl("text")
# note: print is for confirming correct return type, but has addr so not good for testing
# print("txt = %s" % txt)

htm = pkgconflict.HtmlTempl("html")
# print("htm = %s" % htm)

htm2 = pkgconflict.HtmlTemplSame("html2")
# print("htm2 = %s" % htm2)

