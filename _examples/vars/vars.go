// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vars

var (
	V1 = "v1"
	V2 = 42
	V3 = 666.666
)

var (
	V4 string  = "c4"
	V5 int     = 42
	V6 uint    = 42
	V7 float64 = 666.666
)

type Kind int

var (
	Kind1 Kind = 1
	Kind2      = 2
)

// Doc is a top-level string with some documentation attached.
var Doc = "A variable with some documentation"

// FIXME: also use an unexported type
// type kind int
// var (
//  Kind3 kind = 3
//  Kind4      = 4
// )
