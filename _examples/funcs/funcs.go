// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package funcs

import "fmt"

type FunStruct struct {
	FieldI int
	FieldS string
}

func (fs *FunStruct) CallBack(arg1 int, fun func(afs *FunStruct, a1 int, s1 string)) {
	fun(fs, arg1, fs.FieldS)
}

func (fs *FunStruct) CallBackIf(arg1 int, fun func(afs *FunStruct, a1 int, if1 interface{})) {
	fun(fs, arg1, fs.FieldS)
}

func (fs *FunStruct) OtherMeth(arg1 int, args string) {
	fs.FieldI = arg1
	fs.FieldS = args
	fmt.Printf("arg1: %d args: %s\n", arg1, args)
}

// var (
// 	F1 func()
// 	F2 Func
// 	F3 S1
// 	F4 S2
// 	F5 []func()
// 	F6 []Func
// 	F7 [2]func()
// 	F8 [3]Func
// )
//
// type Func func()
//
// type S1 struct {
// 	F1 Func
// 	F2 []Func
// 	F3 [4]Func
// }
//
// type S2 struct {
// 	F1 func()
// 	F2 []func()
// 	F3 [5]func()
// }
//
// func init() {
// 	F1 = func() {
// 		cpkg.Printf("calling F1\n")
// 	}
//
// 	F2 = Func(func() {
// 		cpkg.Printf("calling F2\n")
// 	})
// }
