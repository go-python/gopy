# Copyright 2015 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import structs

print("s = structs.S()")
s = structs.S()
print("s = %s" % (s,))
print("s.Init()")
s.Init()
print("s.Upper('boo')= %s" % repr(s.Upper("boo")).lstrip('u'))

print("s1 = structs.S1()")
s1 = structs.S1()
print("s1 = %s" % (s1,))

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
s2 = structs.S2(1)
print("s2 = %s" % (s2,))

try:
    s2 = structs.S2(1, 2)
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


class S2Child(structs.S2):
    def __init__(self, a, b):
        super(S2Child, self).__init__(a)
        self.local = b

    def __str__(self):
        return ("S2Child{S2: %s, local: %d}"
                % (super(S2Child, self).__str__(), self.local))


try:
    s2child = S2Child(42, 123)
    print("s2child = %s" % (s2child,))
    print("s2child.Public = %s" % (s2child.Public,))
    print("s2child.local = %s" % (s2child.local,))
    print("s2child.private = %s" % (s2child.private,))
except Exception as err:
    print("caught error: %s" % (err,))
    pass

try:
    val = structs.S3()
    val.X = 3
    val.Y = 4
    print("s3.X,Y = %d,%d" % (val.X, val.Y))
except Exception as err:
    print("caught error: %s" % (err,))
    pass

print("OK")
