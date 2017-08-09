// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strconv"
	"strings"
)

// genCdefStruct generates C definitions for a Go struct.
func (g *cffiGen) genCdefStruct(s Struct) {
	g.wrapper.Printf("// A type definition of the %[1]s.%[2]s for wrapping.\n", s.Package().Name(), s.sym.cgoname)
	g.wrapper.Printf("typedef void* %s;\n", s.sym.cgoname)
	g.wrapper.Printf("extern void* cgo_func_%s_new();\n", s.sym.id)
}

// genCdefType generates C definitions for a Go type.
func (g *cffiGen) genCdefType(sym *symbol) {
	g.wrapper.Printf("// C definitions for Go type %[1]s.%[2]s.\n", g.pkg.pkg.Name(), sym.id)
	if !sym.isType() {
		return
	}

	if sym.isStruct() {
		return
	}

	if sym.isBasic() && !sym.isNamed() {
		return
	}

	g.genTypeCdefInit(sym)

	if sym.isNamed() {
		typ := sym.GoType().(*types.Named)
		for imeth := 0; imeth < typ.NumMethods(); imeth++ {
			m := typ.Method(imeth)
			if !m.Exported() {
				continue
			}
			mname := types.ObjectString(m, nil)
			msym := g.pkg.syms.sym(mname)
			if msym == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for [%[1]T] (%#[1]v) (%[2]s)",
					m.Type(),
					m.Name()+" || "+m.FullName(),
				))
			}
			g.genTypeCdefFunc(sym, msym)
		}
	}
	g.genCdefTypeTPStr(sym)
	g.wrapper.Printf("\n")
}

// genTypeCdefInit generates cgo_func_XXXX_new() for a Go type.
func (g *cffiGen) genTypeCdefInit(sym *symbol) {
	if sym.isBasic() {
		btyp := g.pkg.syms.symtype(sym.GoType().Underlying())
		g.wrapper.Printf("typedef %[1]s %[2]s;\n", btyp.cgoname, sym.cgoname)
		g.wrapper.Printf("extern %[1]s cgo_func_%[2]s_new();\n", btyp.cgoname, sym.id)
	} else {
		g.wrapper.Printf("typedef void* %s;\n", sym.cgoname)
		g.wrapper.Printf("extern void* cgo_func_%s_new();\n", sym.id)
	}

	switch {
	case sym.isBasic():
	case sym.isArray():
		typ := sym.GoType().Underlying().(*types.Array)
		elemType := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("extern void cgo_func_%[1]s_ass_item(void* p0, GoInt p1, %[2]s p2);\n", sym.id, elemType.cgoname)
		g.wrapper.Printf("extern %[1]s cgo_func_%[2]s_item(void* p0, %[1]s p1);\n", elemType.cgoname, sym.id)
	case sym.isSlice():
		typ := sym.GoType().Underlying().(*types.Slice)
		elemType := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("extern void cgo_func_%[1]s_ass_item(void* p0, GoInt p1, %[2]s p2);\n", sym.id, elemType.cgoname)
		g.wrapper.Printf("extern void cgo_func_%[1]s_append(void* p0, %[2]s p1);\n", sym.id, elemType.cgoname)
		g.wrapper.Printf("extern %[1]s cgo_func_%[2]s_item(void* p0, %[1]s p1);\n", elemType.cgoname, sym.id)
	case sym.isMap():
		ktyp := sym.GoType().Underlying().(*types.Map).Key()
		etyp := sym.GoType().Underlying().(*types.Map).Elem()
		ksym := g.pkg.syms.symtype(ktyp)
		esym := g.pkg.syms.symtype(etyp)
		g.wrapper.Printf("extern void cgo_func_%[1]s_set(void* p0, %[2]s p1, %[3]s p2);\n", sym.id, ksym.cgoname, esym.cgoname)
		g.wrapper.Printf("extern %[3]s cgo_func_%[1]s_get(void* p0, %[2]s p1);\n", sym.id, ksym.cgoname, esym.cgoname)
		g.wrapper.Printf("extern GoInt cgo_func_%[1]s_len(void* p0);\n", sym.id)
	case sym.isSignature():
	case sym.isInterface():
	default:
		panic(fmt.Errorf(
			"gopy: cdef for %s not handled",
			sym.gofmt(),
		))
	}
}

// genCdefFunc generates a C definition for a Go function.
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

// genTypeCdefFunc generates a C definition for a Go type's method.
func (g *cffiGen) genTypeCdefFunc(sym *symbol, fsym *symbol) {
	if !sym.isType() {
		return
	}

	if sym.isStruct() {
		return
	}

	if sym.isBasic() && !sym.isNamed() {
		return
	}
	isMethod := (sym != nil)
	sig := fsym.GoType().Underlying().(*types.Signature)
	args := sig.Params()
	res := sig.Results()
	var cdef_ret string
	var params []string
	switch res.Len() {
	case 0:
		cdef_ret = "void"
	case 1:
		ret := res.At(0)
		cdef_ret = g.pkg.syms.symtype(ret.Type()).cgoname
	}

	if isMethod {
		params = append(params, sym.cgoname+" p0")
	}

	for i := 0; i < args.Len(); i++ {
		arg := args.At(i)
		index := i
		if isMethod {
			index++
		}
		paramVar := g.pkg.syms.symtype(arg.Type()).cgoname + " " + "p" + strconv.Itoa(index)
		params = append(params, paramVar)
	}
	paramString := strings.Join(params, ", ")
	g.wrapper.Printf("extern %[1]s cgo_func_%[2]s(%[3]s);\n", cdef_ret, fsym.id, paramString)
}

// genCdefMethod generates a C definition for a Go Struct's method.
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

// genCdefStructMemberGetter generates C definitions of Getter/Setter for a Go struct.
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

// genCdefStructMemberSetter generates C defintion of Setter for a Go struct.
func (g *cffiGen) genCdefStructMemberSetter(s Struct, i int, f types.Object) {
	pkg := s.Package()
	ft := f.Type()
	var (
		ifield = newVar(pkg, ft, f.Name(), "ret", "")
	)
	cdef_value := ifield.sym.cgoname
	g.wrapper.Printf("extern void cgo_func_%[1]s_setter_%[2]d(void* p0, %[3]s p1);\n", s.sym.id, i+1, cdef_value)
}

// genCdefStructTPStr generates C definitions of str method for a Go struct.
func (g *cffiGen) genCdefStructTPStr(s Struct) {
	g.wrapper.Printf("extern GoString cgo_func_%[1]s_str(void* p0);\n", s.sym.id)
}

// genCdefTypeTPStr generates C definitions of str method for a Go type.
func (g *cffiGen) genCdefTypeTPStr(sym *symbol) {
	g.wrapper.Printf("extern GoString cgo_func_%[1]s_str(%[2]s p0);\n", sym.id, sym.cgoname)
}
