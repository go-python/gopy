// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/types"
	"strconv"
	"strings"
)

func (g *cffiGen) genCdefStruct(s Struct) {
	g.wrapper.Printf("// A type definition of the %[1]s.%[2]s for wrapping.\n", s.Package().Name(), s.sym.cgoname)
	g.wrapper.Printf("typedef void* %s;\n", s.sym.cgoname)
	g.wrapper.Printf("extern void* cgo_func_%s_new();\n", s.sym.id)
}

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
	var retParams []string
	var cdef_ret string
	sig := f.sig
	rets := sig.Results()
	args := sig.Params()

	switch len(rets) {
	case 0:
		cdef_ret = "void"
	case 1:
		cdef_ret = rets[0].sym.cgoname
	default:
		for i := 0; i < len(rets); i++ {
			retParam := rets[i].sym.cgoname + " r" + strconv.Itoa(i)
			retParams = append(retParams, retParam)
		}
		cdef_ret = "cgo_func_" + f.id + "_return"
		retParamStrings := strings.Join(retParams, "; ")
		g.wrapper.Printf("typedef struct { %[1]s; } %[2]s;\n", retParamStrings, cdef_ret)
	}

	for i := 0; i < len(args); i++ {
		paramVar := args[i].sym.cgoname + " " + "p" + strconv.Itoa(i)
		params = append(params, paramVar)
	}
	paramString := strings.Join(params, ", ")
	g.wrapper.Printf("extern %[1]s cgo_func_%[2]s(%[3]s);\n", cdef_ret, f.id, paramString)
}

func (g *cffiGen) genCdefMethod(f Func) {
	var retParams []string
	var cdef_ret string
	params := []string{"void* p0"}
	sig := f.sig
	rets := sig.Results()
	args := sig.Params()

	switch len(rets) {
	case 0:
		cdef_ret = "void"
	case 1:
		cdef_ret = rets[0].sym.cgoname
	default:
		for i := 0; i < len(rets); i++ {
			retParam := rets[i].sym.cgoname + " r" + strconv.Itoa(i)
			retParams = append(retParams, retParam)
		}
		cdef_ret = "cgo_func_" + f.id + "_return"
		retParamStrings := strings.Join(retParams, "; ")
		g.wrapper.Printf("typedef struct { %[1]s; } %[2]s;\n", retParamStrings, cdef_ret)
	}

	for i := 0; i < len(args); i++ {
		paramVar := args[i].sym.cgoname + " " + "p" + strconv.Itoa(i+1)
		params = append(params, paramVar)
	}
	paramString := strings.Join(params, ", ")
	g.wrapper.Printf("extern %[1]s cgo_func_%[2]s(%[3]s);\n", cdef_ret, f.id, paramString)
}

func (g *cffiGen) genCdefStructMemberGetter(s Struct, i int, f types.Object) {
	pkg := s.Package()
	ft := f.Type()
	var (
		ifield  = newVar(pkg, ft, f.Name(), "ret", "")
		results = []*Var{ifield}
	)
	switch len(results) {
	case 1:
		ret := results[0]
		cdef_ret := ret.sym.cgoname
		g.wrapper.Printf("extern %[1]s cgo_func_%[2]s_getter_%[3]d(void* p0);\n", cdef_ret, s.sym.id, i+1)
	default:
		panic("bind: impossible")
	}
}

func (g *cffiGen) genCdefStructMemberSetter(s Struct, i int, f types.Object) {
	pkg := s.Package()
	ft := f.Type()
	var (
		ifield = newVar(pkg, ft, f.Name(), "ret", "")
	)
	cdef_value := ifield.sym.cgoname
	g.wrapper.Printf("extern void cgo_func_%[1]s_setter_%[2]d(void* p0, %[3]s p1);\n", s.sym.id, i+1, cdef_value)
}

func (g *cffiGen) genCdefStructTPStr(s Struct) {
	g.wrapper.Printf("extern GoString cgo_func_%[1]s_str(void* p0);\n", s.sym.id)
}
