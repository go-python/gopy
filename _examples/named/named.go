// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package named tests various aspects of named types.
package named

// TODO: not dealing with named basic types -- not very intuitive to use
// in python, as they have to become classes.  instead anything like this
// is converted to its basic type when used as an arg / return value
// so you can use them in other methods, but their special type methods
// are not exported

// type Float float32
//
// // Value returns a float32 value
// func (f Float) Value() float32 { return float32(f) }
//
// type X float32
// type XX X
// type XXX XX
// type XXXX XXX
//
// // Value returns a float32 value
// func (x X) Value() float32 { return float32(x) }
//
// // Value returns a float32 value
// func (x XX) Value() float32 { return float32(x) }
//
// // Value returns a float32 value
// func (x XXX) Value() float32 { return float32(x) }
//
// // Value returns a float32 value
// func (x XXXX) Value() float32 { return float32(x) }
//
// type Str string

// Value returns a string value
// func (s Str) Value() string { return string(s) }

type Slice []float64

func (s Slice) At(i int) float64 { return s[i] }

type Array [2]float64

func (a Array) At(i int) float64 { return a[i] }

// type T int
//
// func (t T) PublicMethod()  {}
//func (t T) privateMethod() {}
