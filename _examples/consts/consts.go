// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package consts

const (
	C1 = "c1"
	C2 = 42
	C3 = 666.666
)

const (
	C4 string  = "c4"
	C5 int     = 42
	C6 uint    = 42
	C7 float64 = 666.666
)

type Kind int

const (
	Kind1 Kind = 1
	Kind2      = 2
)

// FIXME: also use an unexported type
// type kind int
// const (
//  Kind3 kind = 3
//  Kind4      = 4
// )
