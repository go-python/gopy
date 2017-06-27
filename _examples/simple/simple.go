// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// simple is a simple package.
package simple

// Func is a simple func
func Func() {}

// Add is a (less) simple func
func Add(i, j int) int {
	return i + j
}

func Bool(b bool) bool {
	if b {
		return false
	}
	return true
}
