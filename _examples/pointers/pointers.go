// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pointers

//type SPtr *S

type S struct {
	Value int
}

// Inc increments the value of s
func Inc(s *S) {
	s.Value++
}

// note: pointers to basic types are not supported -- would
// require a handle -- could to, but probably not worth it..
type MyInt int

// IncInt increments an integer
func IncMyInt(i *MyInt) {
	(*i)++
}

// IncInt increments an integer
func IncInt(i *int) {
	(*i)++
}
