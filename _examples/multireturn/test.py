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

############### Single Str WithoutError Return ##############
oneStrResult = multireturn.SingleStrWithoutErrorFunc()
print("Single Str WithoutError Return %r" % oneStrResult)

############### Single WithError Return ##############
errorFalseResult = multireturn.SingleWithErrorFunc(False)
print("Single WithError(False) Return %r" % errorFalseResult)

try:
	errorTrueResult = multireturn.SingleWithErrorFunc(True)
	print("Failed to throw an exception")
except RuntimeError as ex:
	print("Single WithError(True). Exception:%r" % ex)

############### Double WithoutError Return ##############
twoResults = multireturn.DoubleWithoutErrorFunc()
print("Double WithoutError Return (%r, %r)" % twoResults)

############### Double WithError Return ##############
try:
	value400 = multireturn.DoubleWithErrorFunc(True)
	print("Failed to throw an exception. Return (%r, %r)." % value400)
except RuntimeError as ex:
	print("Double WithError(True). exception Return %r" % ex)

value500 = multireturn.DoubleWithErrorFunc(False)
print("Double WithError(False) Return %r" % value500)

############### Triple Without Error Return ##############
threeResults = multireturn.TripleWithoutErrorFunc1()
print("Triple WithoutError(Without String) Return (%r, %r, %r)" % threeResults)

threeResults = multireturn.TripleWithoutErrorFunc2()
print("Triple WithoutError(With String) Return (%r, %r, %r)" % threeResults)

############### Triple With Error Return ##############
try:
	(value900, value1000, errorTrueResult) = multireturn.TripleWithErrorFunc(True)
except RuntimeError as ex:
	print("Triple WithError(True). exception Return %r" % ex)

(value1100, value1200) = multireturn.TripleWithErrorFunc(False)
print("Triple WithError(False) Return (%r, %r)" % (value1100, value1200))

############### Triple Struct Return With Error ##############
(ptr1300, struct1400, errorTrueResult) = multireturn.TripleWithStructWithErrorFunc(True)
print("Triple WithError(True) Return (%r, %r, %r)" % (ptr1300.P, struct1400.P, errorFalseResult))

(value1500, value1600, errorFalseResult) = multireturn.TripleWithStructWithErrorFunc(False)
print("Triple WithError(False) Return (%r, %r, %r)" % (value1500.P, value1600.P, errorFalseResult))

############### Triple Interface Return Without Error ##############
(interface1700, struct1800, ptr1900) = multireturn.TripleWithInterfaceWithoutErrorFunc()
print("Triple WithError(True) Return (%r, %r, %r)" % (interface1700.P, struct1800.P, ptr1900))
