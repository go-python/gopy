// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"strconv"
	"strings"
)

func (g *cffiGen) genCdef(f Func) {
	var params []string
	var cdef_ret string
	sig := f.sig
	rets := sig.Results()

	switch len(rets) {
	case 0:
		cdef_ret = "void"
	case 1:
		cdef_ret = rets[0].sym.cgoname
	default:
		cdef_ret = "cgo_func_" + f.name + "_return"
	}

	args := sig.Params()
	for i := 0; i < len(args); i++ {
		paramVar := args[i].sym.cgoname + " " + "p" + strconv.Itoa(i)
		params = append(params, paramVar)
	}

	paramString := strings.Join(params, ", ")
	g.builder.Printf("extern %[1]s cgo_func_%[2]s_%[3]s(%[4]s);\n", cdef_ret, g.pkg.Name(), f.name, paramString)
}
