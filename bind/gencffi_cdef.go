// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"strconv"
	"strings"
)

func (g *cffiGen) genCdefType(sym *symbol) {
	if !sym.isType() {
		return
	}

	if sym.isStruct() {
		return
	}

	if sym.isBasic() && !sym.isNamed() {
		return
	}

	if sym.isBasic() {
		btyp := g.pkg.syms.symtype(sym.GoType().Underlying())
		g.wrapper.Printf("typedef %s %s;\n\n", btyp.cgoname, sym.cgoname)
	} else {
		g.wrapper.Printf("typedef void* %s;\n\n", sym.cgoname)
	}
}

func (g *cffiGen) genCdefFunc(f Func) {
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
	g.wrapper.Printf("extern %[1]s cgo_func_%[2]s(%[3]s);\n", cdef_ret, f.id, paramString)
}
