# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
import multireturn, go

############### No Return ##############
noResult = multireturn.NoReturnFunc()
print("No Return %r" % noResult)

############### Single WithoutError Return ##############
oneResult = multireturn.SingleWithoutErrorFunc()
print("Single WithoutError Return %r" % oneResult)

############### Single WithError Return ##############
errorFalseResult = multireturn.SingleWithErrorFunc(False)
print("Single WithError(False) Return %r" % errorFalseResult)

errorTrueResult = multireturn.SingleWithErrorFunc(True)
print("Single WithError(True) Return %r" % errorTrueResult)

############### Double WithoutError Return ##############
twoResults = multireturn.DoubleWithoutErrorFunc()
print("Double WithoutError Return %r" % twoResults)

############### Double WithError Return ##############
(value400, errorTrueResult) = multireturn.DoubleeWithErrorFunc(True)
print("Double WithError(True) Return (%r, %r)" % (value400, errorTrueResult))

(value500, errorFalseResult) = multireturn.DoubleWithErrorFunc(False)
print("Double WithError(False) Return (%r, %r)" % (value500, errorFalseResult))

############### Triple Without Error Return ##############
threeResults = multireturn.TripleWithoutErrorFunc()
print("Triple WithoutError Return %r" % threeResult)

############### Triple With Error Return ##############
(value900, value1000, errorTrueResult) = multireturn.TripleWithErrorFunc(True)
print("Triple WithError(True) Return (%r, %r, %r)" % (value900, value1000, errorFalseResult))

(value1100, value1200, errorFalseResult) = multireturn.TripleWithErrorFunc(False)
print("Triple WithError(False) Return (%r, %r, %r)" % (value1100, value1200, errorFalseResult))

############### Triple Struct Return With Error ##############
(ptr1300, struct1400, errorTrueResult) = multireturn.TripleWithStructWithErrorFunc(True)
print("Triple WithError(True) Return (%r, %r, %r)" % (ptr1300.P, struct1400.P, errorFalseResult))

(value1500, value1600, errorFalseResult) = multireturn.TripleWithStructWithErrorFunc(False)
print("Triple WithError(False) Return (%r, %r, %r)" % (value1500.P, value1600.P, errorFalseResult))

############### Triple Interface Return Without Error ##############
(interface1700, struct1800, ptr1900) = multireturn.TripleWithInterfaceWithoutErrorFunc()
print("Triple WithError(True) Return (%r, %r, %r)" % (interface1700.P, struct1800.P, ptr1900))
