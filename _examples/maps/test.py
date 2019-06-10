# Copyright 2017 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
import maps

a = maps.New()
b = {1: 3.0, 2: 5.0}

# set to true to test all the builtin map functionality
# due to the random nature of map access in Go, this will
# cause go test to fail randomly so it is off by default
# but should be re-enabled when anything significant is
# changed!
testall = False

if testall:
    print('map a:', a)
    print('map a repr:', a.__repr__())

    print('map a keys:', a.keys())
    print('map a values:', a.values())
    print('map a items:', a.items())
    print('map a iter')
    for k,v in a:
        print("key:", k, "value:", v)

print('map a[1]:', a[1])
print('map a[2]:', a[2])
print('2 in map:', 2 in a)
print('3 in map:', 3 in a)

# TODO: not sure why python2 doesn't just catch this error, but it doesn't seem to..
# try:
#     v = a[4]
# except Exception as err:
#     print("caught error: %s" % (err,))
#     pass

    
print('maps.Sum from Go map:', maps.Sum(a))

print('map b:', b)

print('maps.Sum from Python dictionary:', maps.Sum(maps.Map_int_float64(b)))
print('maps.Keys from Go map:', maps.Keys(a))
print('maps.Values from Go map:', maps.Values(a))
print('maps.Keys from Python dictionary:', maps.Keys(maps.Map_int_float64(b)))
print('maps.Values from Python dictionary:', maps.Values(maps.Map_int_float64(b)))

del a[1]
print('deleted 1 from a:', a)

print("OK")
