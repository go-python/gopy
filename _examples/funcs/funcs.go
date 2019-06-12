// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package funcs

import (
	"fmt"

	"github.com/go-python/gopy/_examples/cpkg"
)

type FunStruct struct {
	FieldI int
	FieldS string
}

func (fs *FunStruct) CallBack(i int, fun func(fs *FunStruct, i int, s string)) {
	fun(fs, i, fs.FieldS)
}

type RecvFunc func(fs *FunStruct, i int, v interface{})

func (fs *FunStruct) CallBackIf(i int, fun RecvFunc) {
	fun(fs, i, fs.FieldS)
}

func (fs *FunStruct) CallBackRval(i int, fun func(fs *FunStruct, i int, v interface{}) bool) {
	rv := fun(fs, i, fs.FieldS)
	fmt.Printf("got return value: %v\n", rv)
}

func (fs *FunStruct) OtherMeth(i int, s string) {
	fs.FieldI = i
	fs.FieldS = s
	fmt.Printf("i=%d s=%s\n", i, s)
}

func (fs *FunStruct) ObjArg(ofs *FunStruct) {
	if ofs == nil {
		fmt.Printf("got nil\n")
	} else {
		fmt.Printf("ofs FieldI: %d FieldS: %s\n", ofs.FieldI, ofs.FieldS)
	}
}

var (
	F1 func()
	F2 Func
	F3 S1
	F4 S2
	F5 []func()
	F6 []Func
	F7 [2]func()
	F8 [3]Func
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
	F3 [5]func()
}

func init() {
	F1 = func() {
		cpkg.Printf("calling F1\n")
	}

	F2 = Func(func() {
		cpkg.Printf("calling F2\n")
	})
}
