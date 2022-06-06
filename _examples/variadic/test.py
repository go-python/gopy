# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
import variadic, go

varResult = variadic.VariFunc(1,2,3,4,5)
print("Variadic 1+2+3+4+5 = %d" % varResult)

nonvarResult = variadic.NonVariFunc(1, go.Slice_int([2,3,4]),5)
print("NonVariadic 1+[2+3+4]+5 = %d" % nonvarResult)

if isinstance(varResult, int):
	print("Type OK")
else:
	print("Type Not OK")
