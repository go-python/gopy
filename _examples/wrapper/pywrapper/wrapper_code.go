// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pywrapper

import (
	"fmt"

	"github.com/go-python/gopy/_examples/wrapper"
)

type WrapperStruct struct {
	data *wrapper.RealStruct
}

func Test() string {
	a := wrapper.PointerTest()
	fmt.Println("%t", a)
	return "Hello"
}

func (a WrapperStruct) Test() string {
	return fmt.Sprintf("%t", a.data)
}
