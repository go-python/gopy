# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

## py2/py3 compat
from __future__ import print_function

import threading

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


# Concurrent calls must not swap or drop errors between Python threads: each
# thread's exception (or lack of one) must match what it, specifically, did.
def race():
    n = 200
    bad = []
    lock = threading.Lock()

    def worker(i):
        try:
            if i % 2 == 0:
                pyerrors.Div(10, 0)
                with lock:
                    bad.append((i, "missing exception"))
            else:
                r = pyerrors.Div(10, 1)
                if r != 10:
                    with lock:
                        bad.append((i, "wrong result: %r" % r))
        except Exception as e:
            if i % 2 == 1:
                with lock:
                    bad.append((i, "unexpected exception: %s" % e))
            elif str(e) != "Divide by zero.":
                with lock:
                    bad.append((i, "wrong message: %s" % e))

    threads = [threading.Thread(target=worker, args=(i,)) for i in range(n)]
    for t in threads:
        t.start()
    for t in threads:
        t.join()
    assert not bad, bad


race()
