// Copyright 2020 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package gopygc tests the interaction between the python gc system and gopy
package gopygc

import "bytes"

type StructA struct {
	A int
}

type MapA map[string]string

type SliceA []string

func StructValue() StructA {
	return StructA{
		A: 42,
	}
}

func SliceScalarValue() []int {
	return []int{1, 2}
}

func SliceStructValue() []StructA {
	return []StructA{{1}, {2}}
}

func MapValue() map[int]int {
	return map[int]int{1: 2}
}

func MapValueStruct() map[int]StructA {
	return map[int]StructA{1: StructA{3}}
}

func ExternalType() *bytes.Buffer {
	return &bytes.Buffer{}
}
