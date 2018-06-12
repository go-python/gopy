// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package rename tests changing the names of methods and functions
package rename

//gopy:name say_hi
func SayHi() string {
	return "hi"
}

type MyStruct struct {
}

//gopy:name say_something
func (s *MyStruct) SaySomething() (something string) {
	return "something"
}
