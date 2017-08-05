// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slices

func IntSum(s []int) int {
	sum := 0
	for _, value := range s {
		sum += value
	}
	return sum
}

func CreateSlice() []int {
	return []int{1, 2, 3, 4}
}
