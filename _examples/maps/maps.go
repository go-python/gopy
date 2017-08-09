// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"sort"
)

func Sum(t map[int]float64) float64 {
	sum := 0.0
	for _, v := range t {
		sum += v
	}

	return sum
}

func New() map[int]float64 {
	return map[int]float64{
		1: 3.0,
		2: 5.0,
	}
}

func Keys(t map[int]float64) []int {
	var keys []int
	for k, _ := range t {
		keys = append(keys, k)
	}

	sort.Ints(keys)
	return keys
}

func Values(t map[int]float64) []float64 {
	var values []float64
	for _, v := range t {
		values = append(values, v)
	}

	sort.Float64s(values)
	return values
}
