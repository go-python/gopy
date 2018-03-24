// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Reference from: https://github.com/initialed85/golang_python_binding_research

package lot

type Value struct {
	SomeString        string
	SomeInt           int64
	SomeFloat         float64
	SomeBool          bool
	SomeListOfStrings []string
	SomeListOfInts    []int64
	SomeListOfFloats  []float64
	SomeListOfBools   []bool
}

// New returns a struct with exported fields of different types
func New() Value {
	return Value{
		"some string",
		1337,
		1337.1337,
		true,
		[]string{"some", "list", "of", "strings"},
		[]int64{6, 2, 9, 1},
		[]float64{6.6, 2.2, 9.9, 1.1},
		[]bool{true, false, true, false},
	}
}
