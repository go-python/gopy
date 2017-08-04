# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import structs

print("s = structs.S()")
s = structs.S()
print("s = %s" % (s,))
print("s.Init()")
s.Init()
print("s.Upper('boo')= %r" % (s.Upper("boo"),))

print("s1 = structs.S1()")
s1 = structs.S1()
print("s1 = %s" %(s1,))

try:
    s1 = structs.S1(1)
except Exception as err:
    print("caught error: %s" % (err,))
    pass

try:
    s1 = structs.S1()
    print("s1.private = %s" % (s1.private,))
except Exception as err:
    print("caught error: %s" % (err,))
    pass

print("s2 = structs.S2()")
s2 = structs.S2()
print("s2 = %s" % (s2,))

print("s2 = structs.S2(1)")
s2 = structs.S2(1)
print("s2 = %s" % (s2,))

try:
    s2 = structs.S2(1,2)
except Exception as err:
    print("caught error: %s" % (err,))
    pass

try:
    s2 = structs.S2(42)
    print("s2 = %s" % (s2,))
    print("s2.Public = %s" % (s2.Public,))
    print("s2.private = %s" % (s2.private,))
except Exception as err:
    print("caught error: %s" % (err,))
    pass
