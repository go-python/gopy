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

type SliceUint8 []uint8
type SliceUint16 []uint16
type SliceUint32 []uint32
type SliceUint64 []uint64

type SliceInt8 []int8
type SliceInt16 []int16
type SliceInt32 []int32
type SliceInt64 []int64
