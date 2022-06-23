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
errorFalseResult1 = multireturn.SingleWithErrorFunc(False)
print("Single WithError(False) Return %r" % errorFalseResult1)

try:
	errorTrueResult1 = multireturn.SingleWithErrorFunc(True)
	print("Failed to throw an exception")
except RuntimeError as ex:
	print("Single WithError(True). Exception: %r" % ex)

############### Double WithoutError Return ##############
twoResults = multireturn.DoubleWithoutErrorFunc1()
print("Double WithoutError(Without String) Return (%r, %r)" % twoResults)

twoResults = multireturn.DoubleWithoutErrorFunc2()
print("Double WithoutError(With String) Return (%r, %r)" % twoResults)

############### Double WithError Return ##############
try:
	value400 = multireturn.DoubleWithErrorFunc(True)
	print("Failed to throw an exception. Return (%r, %r)." % value400)
except RuntimeError as ex:
	print("Double WithError(True). Exception: %r" % ex)

value500 = multireturn.DoubleWithErrorFunc(False)
print("Double WithError(False) Return %r" % value500)

############### Triple Without Error Return ##############
threeResults = multireturn.TripleWithoutErrorFunc1()
print("Triple WithoutError(Without String) Return (%r, %r, %r)" % threeResults)

threeResults = multireturn.TripleWithoutErrorFunc2()
print("Triple WithoutError(With String) Return (%r, %r, %r)" % threeResults)

############### Triple With Error Return ##############
try:
	(value900, value1000) = multireturn.TripleWithErrorFunc(True)
	print("Triple WithError(True) Return (%r, %r, %r)" % (value900, value1000))
except RuntimeError as ex:
	print("Triple WithError(True) Exception: %r" % ex)

(value1100, value1200) = multireturn.TripleWithErrorFunc(False)
print("Triple WithError(False) Return (%r, %r)" % (value1100, value1200))

############### Triple Struct Return With Error ##############
try:
	(ptr1300, struct1400) = multireturn.TripleWithStructWithErrorFunc(True)
	print("Triple WithError(True) Return (%r, %r)" % (ptr1300.P, struct1400.P))
except RuntimeError as ex:
	print("Triple WithError(True) Exception: %r" % ex)

(value1500, value1600) = multireturn.TripleWithStructWithErrorFunc(False)
print("Triple WithError(False) Return (%r, %r)" % (value1500.P, value1600.P))

############### Triple Interface Return Without Error ##############
(interface1700, struct1800, ptr1900) = multireturn.TripleWithInterfaceWithoutErrorFunc()
print("Triple WithoutError() Return (%r, %r, %r)" % (interface1700.Number(), struct1800.P, ptr1900.P))

############## Function Returning Functions ignored ##############
assert("FunctionReturningFunction" not in dir(multireturn))
