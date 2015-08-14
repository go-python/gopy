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
type T struct{}

func (t *T) F() {
	cpkg.Printf("t.F [CALLED]\n")
}

// CallIface calls F() on v
func CallIface(v Iface) {
	cpkg.Printf("iface.CallIface...\n")
	v.F()
	cpkg.Printf("iface.CallIface... [DONE]\n")
}
