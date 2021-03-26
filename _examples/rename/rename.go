// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package rename tests changing the names of methods and functions
package rename

// gopy:name say_hi_fn
// Comment follows the tag and function should be renamed
// to say_hi_fn
func SayHi() string {
	return "hi"
}

// I should be renamed to auto_renamed_func, when generated
// with -rename flag
func AutoRenamedFunc() {

}

// MyStruct has two fields
type MyStruct struct {
	// I should be renamed to auto_renamed_property
	// when generated with -rename flag
	AutoRenamedProperty string

	// I should be renamed to custom_name with the custom option
	AutoRenamedProperty2 string `gopy:"custom_name"`
}

// A method that says something
// gopy:name say_something
func (s *MyStruct) SaySomethingFunc() (something string) {
	return "something"
}

// I should be renamed to auto_renamed_meth, when generated
// with -rename flag
func (s *MyStruct) AutoRenamedMeth() {

}
