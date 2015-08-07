// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package funcs

var (
	F1 func()
	F2 Func
	F3 S1
	F4 S2
	F5 []func()
	F6 []Func
	F7 [2]func()
	F7 [2]Func
)

type Func func()

type S1 struct {
	F1 Func
	F2 []Func
	F3 [4]Func
}

type S2 struct {
	F1 func()
	F2 []func()
	F3 [4]func()
}
