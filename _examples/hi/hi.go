// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package hi exposes a few Go functions to be wrapped and used from Python.
package hi

import (
	"fmt"
)

// Hi prints hi from Go
func Hi() {
	fmt.Printf("hi from go\n")
}

// Hello prints a greeting from Go
func Hello(s string) {
	fmt.Printf("hello %s from go\n", s)
}

// Concat concatenates two strings together and returns the resulting string.
func Concat(s1, s2 string) string {
	return s1 + s2
}

// Add returns the sum of its arguments.
func Add(i, j int) int {
	return i + j
}
