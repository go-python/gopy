// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package sliceptr tests support for pointer-to-slice arguments
// useful for getting bulk data from Go to Python

package sliceptr

import "strconv"

type StrVector []string
type IntVector []int

func Fill(iv *IntVector) {
	*iv = IntVector{1, 2, 3}
}

func Append(iv *IntVector) {
	*iv = append(*iv, 4)
}

func Convert(iv *IntVector, sv *StrVector) {
	for _, v := range *iv {
		*sv = append(*sv, strconv.Itoa(v))
	}
}

// compiles, but dies with SIGSEGV
//func (iv *IntVector) Convert(sv *StrVector) {
//	for _, v := range *iv {
//		*sv = append(*sv, strconv.Itoa(v))
//	}
//}
