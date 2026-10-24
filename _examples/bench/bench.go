// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bench has a few trivial functions used to compare the per-call
// overhead of gopy's backends (GOPY_BACKEND); see bench.py and the
// "benchmark" job in .github/workflows/ci.yml.  Not part of the test suite
// itself: no _examples/bench entry in main_test.go's features map.
package bench

// Add returns the sum of its arguments: the cheapest possible call, to
// isolate per-call FFI overhead from any argument-marshaling cost.
func Add(i, j int) int {
	return i + j
}

// Concat concatenates two strings: a second data point, since string
// arguments/returns cross the boundary very differently from an int
// (a managed buffer + length or a null-terminated copy, depending on the
// backend) and might not share the int case's relative cost.
func Concat(a, b string) string {
	return a + b
}
