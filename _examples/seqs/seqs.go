// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package seqs tests various aspects of sequence types.
package seqs

type Slice []float64

func (s Slice) At(i int) float64 { return s[i] }

type Array [10]float64

func (a Array) At(i int) float64 { return a[i] }
