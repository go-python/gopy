// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wrapper

type RealStruct struct {
	pointers map[string]*RealStruct
	Channel  chan int
}

func PointerTest() *RealStruct {
	return &RealStruct{}
}
