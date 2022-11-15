// Copyright 2022 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package variadic

// ///////////// Non Variadic //////////////
func NonVariFunc(arg1 int, arg2 []int, arg3 int) int {
	total := arg1
	for _, num := range arg2 {
		total += num
	}
	total += arg3

	return total
}

// ///////////// Variadic Over Int //////////////
func VariFunc(vargs ...int) int {
	total := 0
	for _, num := range vargs {
		total += num
	}
	return total
}

// ///////////// Variadic Over Struct //////////////
type IntStrUct struct {
	p int
}

func NewIntStrUct(n int) IntStrUct {
	return IntStrUct{
		p: n,
	}
}

func VariStructFunc(vargs ...IntStrUct) int {
	total := 0
	for _, inst := range vargs {
		total += inst.p
	}
	return total
}

// ///////////// Variadic Over Interface //////////////
type IntInterFace interface {
	Number() int
}

func (is *IntStrUct) Number() int {
	return is.p
}

func VariInterFaceFunc(vargs ...IntInterFace) int {
	total := 0
	for _, inst := range vargs {
		total += inst.Number()
	}
	return total
}
