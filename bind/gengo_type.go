// Copyright 2016 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
)

func (g *goGen) genStruct(s Struct) {
	//fmt.Printf("obj: %#v\ntyp: %#v\n", obj, typ)
	typ := s.Struct()
	g.Printf("\n// --- wrapping %s ---\n\n", s.sym.gofmt())

	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() {
			continue
		}

		ft := f.Type()
		fsym := g.pkg.syms.symtype(ft)
		ftname := fsym.cgotypename()
		if needWrapType(ft) {
			ftname = fmt.Sprintf("cgo_type_%[1]s_field_%d", s.ID(), i+1)
			g.Printf("//export %s\n", ftname)
			g.Printf("type %s unsafe.Pointer\n\n", ftname)
		}

		// -- getter --

		g.Printf("//export cgo_func_%[1]s_getter_%[2]d\n", s.ID(), i+1)
		g.Printf("func cgo_func_%[1]s_getter_%[2]d(self cgo_type_%[1]s) %[3]s {\n",
			s.ID(), i+1,
			ftname,
		)
		g.Indent()
		g.Printf(
			"ret := (*%[1]s)(unsafe.Pointer(self))\n",
			s.sym.gofmt(),
		)

		if !fsym.isBasic() {
			g.Printf("cgopy_incref(unsafe.Pointer(&ret.%s))\n", f.Name())
			g.Printf("return %s(unsafe.Pointer(&ret.%s))\n", ftname, f.Name())
		} else {
			g.Printf("return ret.%s\n", f.Name())
		}
		g.Outdent()
		g.Printf("}\n\n")

		// -- setter --
		g.Printf("//export cgo_func_%[1]s_setter_%[2]d\n", s.ID(), i+1)
		g.Printf("func cgo_func_%[1]s_setter_%[2]d(self cgo_type_%[1]s, v %[3]s) {\n",
			s.ID(), i+1, ftname,
		)
		g.Indent()
		fset := "v"
		if !fsym.isBasic() {
			fset = fmt.Sprintf("*(*%s)(unsafe.Pointer(v))", fsym.gofmt())
		}
		g.Printf(
			"(*%[1]s)(unsafe.Pointer(self)).%[2]s = %[3]s\n",
			s.sym.gofmt(),
			f.Name(),
			fset,
		)
		g.Outdent()
		g.Printf("}\n\n")
	}

	for _, m := range s.meths {
		g.genMethod(s, m)
	}

	g.genFuncNew(s.fnew, s, s.sym)
	g.genFunc(s.fnew)

	/*
		// empty interface converter
		g.Printf("//export cgo_func_%[1]s_eface\n", s.ID())
		g.Printf("func cgo_func_%[1]s_eface(self %[2]s) interface{} {\n",
			s.sym.id,
			s.sym.cgoname,
		)
		g.Indent()
		g.Printf("var v interface{} = ")
		if s.sym.isBasic() {
			g.Printf("%[1]s(self)\n", s.sym.gofmt())
		} else {
			g.Printf("*(*%[1]s)(unsafe.Pointer(self))\n", s.sym.gofmt())
		}
		g.Printf("return v\n")
		g.Outdent()
		g.Printf("}\n\n")
	*/

	// support for __str__
	g.genFuncTPStr(s, s.sym, s.prots&ProtoStringer == 1)
}

func (g *goGen) genMethod(s Struct, m Func) {
	sig := m.Signature()
	params := "(self cgo_type_" + s.ID()
	if len(sig.Params()) > 0 {
		params += ", " + g.tupleString(sig.Params())
	}
	params += ")"
	ret := " (" + g.tupleString(sig.Results()) + ") "

	g.Printf("//export cgo_func_%[1]s\n", m.ID())
	g.Printf("func cgo_func_%[1]s%[2]s%[3]s{\n",
		m.ID(),
		params,
		ret,
	)
	g.Indent()
	g.genMethodBody(s, m)
	g.Outdent()
	g.Printf("}\n\n")
}

func (g *goGen) genMethodBody(s Struct, m Func) {
	sig := m.Signature()
	results := sig.Results()
	for i := range results {
		if i > 0 {
			g.Printf(", ")
		}
		g.Printf("_gopy_%03d", i)
	}
	if len(results) > 0 {
		g.Printf(" := ")
	}

	g.Printf("(*%s)(unsafe.Pointer(self)).%s(",
		s.sym.gofmt(),
		m.GoName(),
	)

	args := sig.Params()
	for i, arg := range args {
		tail := ""
		if i+1 < len(args) {
			tail = ", "
		}
		if arg.sym.isStruct() {
			g.Printf("*(*%s)(unsafe.Pointer(%s))%s", arg.sym.gofmt(), arg.Name(), tail)
		} else {
			g.Printf("%s%s", arg.Name(), tail)
		}
	}
	g.Printf(")\n")

	if len(results) <= 0 {
		return
	}

	g.Printf("return ")
	for i, res := range results {
		if i > 0 {
			g.Printf(", ")
		}
		// if needWrap(res.GoType()) {
		// 	g.Printf("")
		// }
		if res.needWrap() {
			g.Printf("%s(unsafe.Pointer(&", res.sym.cgoname)
		}
		g.Printf("_gopy_%03d", i)
		if res.needWrap() {
			g.Printf("))")
		}
	}
	g.Printf("\n")

}

func (g *goGen) genType(sym *symbol) {
	if !sym.isType() {
		return
	}
	if sym.isStruct() {
		return
	}
	if sym.isBasic() && !sym.isNamed() {
		return
	}

	g.Printf("\n// --- wrapping %s ---\n\n", sym.gofmt())
	g.Printf("//export %[1]s\n", sym.cgoname)
	g.Printf("// %[1]s wraps %[2]s\n", sym.cgoname, sym.gofmt())
	if sym.isBasic() {
		// we need to reach at the underlying type
		btyp := sym.GoType().Underlying().String()
		g.Printf("type %[1]s %[2]s\n\n", sym.cgoname, btyp)
	} else {
		g.Printf("type %[1]s unsafe.Pointer\n\n", sym.cgoname)
	}
	g.Printf("//export cgo_func_%[1]s_new\n", sym.id)
	g.Printf("func cgo_func_%[1]s_new() %[2]s {\n", sym.id, sym.cgoname)
	g.Indent()
	g.Printf("var o %[1]s\n", sym.gofmt())
	if sym.isBasic() {
		g.Printf("return %[1]s(o)\n", sym.cgoname)
	} else {
		g.Printf("cgopy_incref(unsafe.Pointer(&o))\n")
		g.Printf("return (%[1]s)(unsafe.Pointer(&o))\n", sym.cgoname)
	}
	g.Outdent()
	g.Printf("}\n\n")

	// empty interface converter
	g.Printf("//export cgo_func_%[1]s_eface\n", sym.id)
	g.Printf("func cgo_func_%[1]s_eface(self %[2]s) interface{} {\n",
		sym.id,
		sym.cgoname,
	)
	g.Indent()
	g.Printf("var v interface{} = ")
	if sym.isBasic() {
		g.Printf("%[1]s(self)\n", sym.gofmt())
	} else {
		g.Printf("*(*%[1]s)(unsafe.Pointer(self))\n", sym.gofmt())
	}
	g.Printf("return v\n")
	g.Outdent()
	g.Printf("}\n\n")

	// support for __str__
	g.Printf("// cgo_func_%[1]s_str wraps Stringer\n", sym.id)
	g.Printf(
		"func cgo_func_%[1]s_str(out, in *seq.Buffer) {\n",
		sym.id,
	)
	g.Indent()
	g.genRead("o", "in", sym.GoType())
	g.Printf("str := fmt.Sprintf(\"%%#v\", o)\n")
	g.Printf("out.WriteString(str)\n")
	g.Outdent()
	g.Printf("}\n\n")

	if sym.isArray() || sym.isSlice() {
		var etyp types.Type
		switch typ := sym.GoType().(type) {
		case *types.Array:
			etyp = typ.Elem()
		case *types.Slice:
			etyp = typ.Elem()
		case *types.Named:
			switch typ := typ.Underlying().(type) {
			case *types.Array:
				etyp = typ.Elem()
			case *types.Slice:
				etyp = typ.Elem()
			default:
				panic(fmt.Errorf("gopy: unhandled type [%#v]", typ))
			}
		default:
			panic(fmt.Errorf("gopy: unhandled type [%#v]", typ))
		}
		esym := g.pkg.syms.symtype(etyp)
		if esym == nil {
			panic(fmt.Errorf("gopy: could not retrieve element type of %#v",
				sym,
			))
		}

		// support for __getitem__
		g.Printf("//export cgo_func_%[1]s_item\n", sym.id)
		g.Printf(
			"func cgo_func_%[1]s_item(self %[2]s, i int) %[3]s {\n",
			sym.id,
			sym.cgoname,
			esym.cgotypename(),
		)
		g.Indent()
		g.Printf("arr := (*%[1]s)(unsafe.Pointer(self))\n", sym.gofmt())
		g.Printf("elt := (*arr)[i]\n")
		if !esym.isBasic() {
			g.Printf("cgopy_incref(unsafe.Pointer(&elt))\n")
			g.Printf("return (%[1]s)(unsafe.Pointer(&elt))\n", esym.cgotypename())
		} else {
			if esym.isNamed() {
				g.Printf("return %[1]s(elt)\n", esym.cgotypename())
			} else {
				g.Printf("return elt\n")
			}
		}
		g.Outdent()
		g.Printf("}\n\n")

		// support for __setitem__
		g.Printf("//export cgo_func_%[1]s_ass_item\n", sym.id)
		g.Printf("func cgo_func_%[1]s_ass_item(self %[2]s, i int, v %[3]s) {\n",
			sym.id,
			sym.cgoname,
			esym.cgotypename(),
		)
		g.Indent()
		g.Printf("arr := (*%[1]s)(unsafe.Pointer(self))\n", sym.gofmt())
		g.Printf("(*arr)[i] = ")
		if !esym.isBasic() {
			g.Printf("*(*%[1]s)(unsafe.Pointer(v))\n", esym.gofmt())
		} else {
			if esym.isNamed() {
				g.Printf("%[1]s(v)\n", esym.gofmt())
			} else {
				g.Printf("v\n")
			}
		}
		g.Outdent()
		g.Printf("}\n\n")
	}

	if sym.isSlice() {
		etyp := sym.GoType().Underlying().(*types.Slice).Elem()
		esym := g.pkg.syms.symtype(etyp)
		if esym == nil {
			panic(fmt.Errorf("gopy: could not retrieve element type of %#v",
				sym,
			))
		}

		// support for __append__
		g.Printf("//export cgo_func_%[1]s_append\n", sym.id)
		g.Printf("func cgo_func_%[1]s_append(self %[2]s, v %[3]s) {\n",
			sym.id,
			sym.cgoname,
			esym.cgotypename(),
		)
		g.Indent()
		g.Printf("slice := (*%[1]s)(unsafe.Pointer(self))\n", sym.gofmt())
		g.Printf("*slice = append(*slice, ")
		if !esym.isBasic() {
			g.Printf("*(*%[1]s)(unsafe.Pointer(v))", esym.gofmt())
		} else {
			if esym.isNamed() {
				g.Printf("%[1]s(v)", esym.gofmt())
			} else {
				g.Printf("v")
			}
		}
		g.Printf(")\n")
		g.Outdent()
		g.Printf("}\n\n")
	}

	g.genTypeTPCall(sym)

	g.genTypeMethods(sym)

}

func (g *goGen) genTypeTPCall(sym *symbol) {
	if !sym.isSignature() {
		return
	}

	sig := sym.GoType().Underlying().(*types.Signature)
	if sig.Recv() != nil {
		// don't generate tp_call for methods.
		return
	}

	// support for __call__
	g.Printf("//export cgo_func_%[1]s_call\n", sym.id)
	g.Printf("func cgo_func_%[1]s_call(self %[2]s", sym.id, sym.cgotypename())
	params := sig.Params()
	res := sig.Results()
	if params != nil && params.Len() > 0 {
		for i := 0; i < params.Len(); i++ {
			arg := params.At(i)
			sarg := g.pkg.syms.symtype(arg.Type())
			if sarg == nil {
				panic(fmt.Errorf(
					"gopy: could not find symtype for [%T]",
					arg.Type(),
				))
			}
			g.Printf(", arg%03d %s", i, sarg.cgotypename())
		}
	}
	g.Printf(")")
	if res != nil && res.Len() > 0 {
		g.Printf(" (")
		for i := 0; i < res.Len(); i++ {
			ret := res.At(i)
			sret := g.pkg.syms.symtype(ret.Type())
			if sret == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for [%T]",
					ret.Type(),
				))
			}
			comma := ", "
			if i == 0 {
				comma = ""
			}
			g.Printf("%s%s", comma, sret.cgotypename())
		}
		g.Printf(")")
	}
	g.Printf(" {\n")
	g.Indent()
	if res != nil && res.Len() > 0 {
		for i := 0; i < res.Len(); i++ {
			if i > 0 {
				g.Printf(", ")
			}
			g.Printf("res%03d", i)
		}
		g.Printf(" := ")
	}
	g.Printf("(*(*%[1]s)(unsafe.Pointer(self)))(", sym.gofmt())
	if params != nil && params.Len() > 0 {
		for i := 0; i < params.Len(); i++ {
			comma := ", "
			if i == 0 {
				comma = ""
			}
			arg := params.At(i)
			sarg := g.pkg.syms.symtype(arg.Type())
			if sarg.isBasic() {
				g.Printf("%sarg%03d", comma, i)
			} else {
				g.Printf(
					"%s*(*%s)(unsafe.Pointer(arg%03d))",
					comma,
					sarg.gofmt(),
					i,
				)
			}
		}
	}
	g.Printf(")\n")
	if res != nil && res.Len() > 0 {
		for i := 0; i < res.Len(); i++ {
			ret := res.At(i)
			sret := g.pkg.syms.symtype(ret.Type())
			if !needWrapType(sret.GoType()) {
				continue
			}
			g.Printf("cgopy_incref(unsafe.Pointer(&arg%03d))", i)
		}

		g.Printf("return ")
		for i := 0; i < res.Len(); i++ {
			if i > 0 {
				g.Printf(", ")
			}
			ret := res.At(i)
			sret := g.pkg.syms.symtype(ret.Type())
			if needWrapType(ret.Type()) {
				g.Printf("%s(unsafe.Pointer(&", sret.cgotypename())
			}
			g.Printf("res%03d", i)
			if needWrapType(ret.Type()) {
				g.Printf("))")
			}
		}
		g.Printf("\n")
	}
	g.Outdent()
	g.Printf("}\n\n")

}

func (g *goGen) genTypeMethods(sym *symbol) {
	if !sym.isNamed() {
		return
	}

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
		g.Printf("//export cgo_func_%[1]s\n", msym.id)
		g.Printf("func cgo_func_%[1]s(self %[2]s",
			msym.id,
			sym.cgoname,
		)
		sig := m.Type().(*types.Signature)
		params := sig.Params()
		if params != nil {
			for i := 0; i < params.Len(); i++ {
				arg := params.At(i)
				sarg := g.pkg.syms.symtype(arg.Type())
				if sarg == nil {
					panic(fmt.Errorf(
						"gopy: could not find symbol for [%T]",
						arg.Type(),
					))
				}
				g.Printf(", arg%03d %s", i, sarg.cgotypename())
			}
		}
		g.Printf(") ")
		res := sig.Results()
		if res != nil {
			g.Printf("(")
			for i := 0; i < res.Len(); i++ {
				if i > 0 {
					g.Printf(", ")
				}
				ret := res.At(i)
				sret := g.pkg.syms.symtype(ret.Type())
				if sret == nil {
					panic(fmt.Errorf(
						"gopy: could not find symbol for [%T]",
						ret.Type(),
					))
				}
				g.Printf("%s", sret.cgotypename())
			}
			g.Printf(")")
		}
		g.Printf(" {\n")
		g.Indent()

		if res != nil {
			for i := 0; i < res.Len(); i++ {
				if i > 0 {
					g.Printf(", ")
				}
				g.Printf("res%03d", i)
			}
			if res.Len() > 0 {
				g.Printf(" := ")
			}
		}
		if sym.isBasic() {
			g.Printf("(*%s)(unsafe.Pointer(&self)).%s(",
				sym.gofmt(),
				msym.goname,
			)
		} else {
			g.Printf("(*%s)(unsafe.Pointer(self)).%s(",
				sym.gofmt(),
				msym.goname,
			)
		}

		if params != nil {
			for i := 0; i < params.Len(); i++ {
				if i > 0 {
					g.Printf(", ")
				}
				sarg := g.pkg.syms.symtype(params.At(i).Type())
				if needWrapType(sarg.GoType()) {
					g.Printf("*(*%s)(unsafe.Pointer(arg%03d))",
						sarg.gofmt(),
						i,
					)
				} else {
					g.Printf("arg%03d", i)
				}
			}
		}
		g.Printf(")\n")

		if res == nil || res.Len() <= 0 {
			g.Outdent()
			g.Printf("}\n\n")
			continue
		}

		g.Printf("return ")
		for i := 0; i < res.Len(); i++ {
			if i > 0 {
				g.Printf(", ")
			}
			sret := g.pkg.syms.symtype(res.At(i).Type())
			if needWrapType(sret.GoType()) {
				g.Printf(
					"%s(unsafe.Pointer(&",
					sret.cgoname,
				)
			}
			g.Printf("res%03d", i)
			if needWrapType(sret.GoType()) {
				g.Printf("))")
			}
		}
		g.Printf("\n")

		g.Outdent()
		g.Printf("}\n\n")
	}
}
