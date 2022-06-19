# Copyright 2018 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
import variadic, go

############### Non Variadic ##############
nonvarResult = variadic.NonVariFunc(1, go.Slice_int([2,3,4]),5)
print("NonVariadic 1+[2+3+4]+5 = %d" % nonvarResult)

############### Variadic Over Int ##############
varResult = variadic.VariFunc(1,2,3,4,5)
print("Variadic 1+2+3+4+5 = %d" % varResult)

############### Variadic Over Struct ##############
varStructResult = variadic.VariStructFunc(variadic.NewIntStrUct(1), variadic.NewIntStrUct(2), variadic.NewIntStrUct(3))
print("Variadic Struct s(1)+s(2)+s(3) = %d" % varStructResult)

############### Variadic Over InterFace ##############
varInterFaceResult = variadic.VariInterFaceFunc(variadic.NewIntStrUct(1), variadic.NewIntStrUct(2), variadic.NewIntStrUct(3))
print("Variadic InterFace i(1)+i(2)+i(3) = %d" % varInterFaceResult)

############### Final ##############
if isinstance(varResult, int):
	print("Type OK")
else:
	print("Type Not OK")
