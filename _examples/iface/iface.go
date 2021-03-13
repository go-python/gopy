// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package iface tests various aspects of interfaces.
package iface

import (
	"github.com/go-python/gopy/_examples/cpkg"
)

// Iface has a single F() method
type Iface interface {
	F()
}

// T implements Iface
type T struct {
	FldI int
	FldS string
}

func (t *T) F() {
	cpkg.Printf("t.F [CALLED]\n")
}

// CallIface calls F() on v
func CallIface(v Iface) {
	cpkg.Printf("iface.CallIface...\n")
	v.F()
	cpkg.Printf("iface.CallIface... [DONE]\n")
}

// by default, interface{} is converted to string (most universal type)
func IfaceString(str interface{}) {
	cpkg.Printf("iface as string: %v\n", str)
}

// gopy:interface=handle
// this magic directive says, treat the interface arg as a handle
func IfaceHandle(ifc interface{}) {
	cpkg.Printf("iface as handle: %v\n", ifc)
}
