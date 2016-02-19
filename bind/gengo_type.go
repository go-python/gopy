// Copyright 2016 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/token"
	"go/types"
)

func (g *goGen) genStruct(s Type) {
	//fmt.Printf("obj: %#v\ntyp: %#v\n", obj, typ)
	typ := s.Struct()
	g.Printf("\n// --- wrapping %s ---\n\n", s.gofmt())

	recv := types.NewVar(token.NoPos, s.Package().pkg, "recv", s.GoType())

	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() {
			continue
		}

		ft := f.Type()

		// -- getter --
		fget := Func{
			pkg:  s.pkg,
			sig:  types.NewSignature(recv, nil, types.NewTuple(f), false),
			typ:  nil,
			name: f.Name(),
			desc: s.Descriptor() + "." + f.Name() + ".get",
			id:   s.ID() + "_" + f.Name() + "_get",
			doc:  "",
			ret:  ft,
			err:  false,
		}
		g.genFuncGetter(fget, s)
		g.genMethod(s, fget)

		// -- setter --
		fset := Func{
			pkg:  s.pkg,
			sig:  types.NewSignature(recv, types.NewTuple(f), nil, false),
			typ:  nil,
			name: f.Name(),
			desc: s.Descriptor() + "." + f.Name() + ".set",
			id:   s.ID() + "_" + f.Name() + "_set",
			doc:  "",
			ret:  nil,
			err:  false,
		}
		g.genFuncSetter(fset, s)
		g.genMethod(s, fset)
	}

	for _, m := range s.meths {
		g.genMethod(s, m)
	}

	g.genFuncNew(s.funcs.new, s)
	g.genFunc(s.funcs.new)

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
	g.genFuncTPStr(s)
	g.genMethod(s, s.funcs.str)
}

func (g *goGen) genMethod(s Type, m Func) {
	g.Printf("\n// cgo_func_%[1]s wraps %[2]s.%[3]s\n",
		m.ID(),
		s.gofmt(), m.GoName(),
	)
	g.Printf("func cgo_func_%[1]s(out, in *seq.Buffer) {\n", m.ID())
	g.Indent()
	g.genMethodBody(s, m)
	g.Outdent()
	g.Printf("}\n\n")

	g.regs = append(g.regs, goReg{
		Descriptor: m.Descriptor(),
		ID:         uhash(m.ID()),
		Func:       m.ID(),
	})
}

func (g *goGen) genMethodBody(s Type, m Func) {
	g.genRead("o", "in", s.GoType())

	sig := m.Signature()
	args := sig.Params()
	if args != nil {
		for i := 0; i < args.Len(); i++ {
			arg := args.At(i)
			g.Printf("// arg-%03d: %v\n", i, gofmt(g.pkg.Name(), arg.Type()))
			g.genRead(fmt.Sprintf("_arg_%03d", i), "in", arg.Type())
		}
	}

	results := sig.Results()
	if results != nil {
		for i := 0; i < results.Len(); i++ {
			if i > 0 {
				g.Printf(", ")
			}
			g.Printf("_res_%03d", i)
		}
		g.Printf(" := ")
	}

	if m.typ == nil {
		src := s.GoType() // FIXME(sbinet)
		cnv := g.cnv(src, src, "o")
		g.Printf("cgo_func_%s_(%s", m.ID(), cnv)
		if args != nil && args.Len() > 0 {
			g.Printf(", ")
		}
	} else {
		g.Printf("o.%s(", m.GoName())
	}

	if args != nil {
		for i := 0; i < args.Len(); i++ {
			tail := ""
			if i+1 < args.Len() {
				tail = ", "
			}
			arg := args.At(i)
			switch typ := arg.Type().Underlying().(type) {
			case *types.Struct:
				ptr := types.NewPointer(typ)
				g.Printf("%s%s", g.cnv(typ, ptr, fmt.Sprintf("_arg_%03d", i)), tail)
			default:
				g.Printf("_arg_%03d%s", i, tail)
			}
		}
	}
	g.Printf(")\n")

	if results == nil {
		return
	}

	for i := 0; i < results.Len(); i++ {
		res := results.At(i)
		g.genWrite(fmt.Sprintf("_res_%03d", i), "out", res.Type())
	}
}

func (g *goGen) genType(typ Type) {
	if typ.isStruct() {
		g.genStruct(typ)
		return
	}
	if typ.isBasic() && !typ.isNamed() {
		return
	}

	g.Printf("\n// --- wrapping %s ---\n\n", typ.gofmt())

	g.genFuncNew(typ.funcs.new, typ)
	g.genFunc(typ.funcs.new)

	/*
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
	*/

	// support for __str__
	g.genFuncTPStr(typ)
	g.genMethod(typ, typ.funcs.str)

	if typ.isArray() || typ.isSlice() {
		var etyp types.Type
		switch t := typ.GoType().(type) {
		case *types.Array:
			etyp = t.Elem()
		case *types.Slice:
			etyp = t.Elem()
		case *types.Named:
			switch tn := t.Underlying().(type) {
			case *types.Array:
				etyp = tn.Elem()
			case *types.Slice:
				etyp = tn.Elem()
			default:
				panic(fmt.Errorf("gopy: unhandled type [%#v]", typ))
			}
		default:
			panic(fmt.Errorf("gopy: unhandled type [%#v]", typ))
		}
		esym := g.pkg.syms.symtype(etyp)
		if esym == nil {
			panic(fmt.Errorf("gopy: could not retrieve element type of %#v",
				typ,
			))
		}

		// support for __getitem__
		g.Printf("//export cgo_func_%[1]s_item\n", typ.ID())
		g.Printf(
			"func cgo_func_%[1]s_item(self %[2]s, i int) %[3]s {\n",
			typ.ID(),
			typ.CGoName(),
			esym.cgotypename(),
		)
		g.Indent()
		g.Printf("arr := (*%[1]s)(unsafe.Pointer(self))\n", typ.gofmt())
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
		g.Printf("//export cgo_func_%[1]s_ass_item\n", typ.ID())
		g.Printf("func cgo_func_%[1]s_ass_item(self %[2]s, i int, v %[3]s) {\n",
			typ.ID(),
			typ.CGoName(),
			esym.cgotypename(),
		)
		g.Indent()
		g.Printf("arr := (*%[1]s)(unsafe.Pointer(self))\n", typ.gofmt())
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

	if typ.isSlice() {
		etyp := typ.GoType().Underlying().(*types.Slice).Elem()
		esym := g.pkg.syms.symtype(etyp)
		if esym == nil {
			panic(fmt.Errorf("gopy: could not retrieve element type of %#v",
				typ,
			))
		}

		// support for __append__
		g.Printf("//export cgo_func_%[1]s_append\n", typ.ID())
		g.Printf("func cgo_func_%[1]s_append(self %[2]s, v %[3]s) {\n",
			typ.ID(),
			typ.CGoName(),
			esym.cgotypename(),
		)
		g.Indent()
		g.Printf("slice := (*%[1]s)(unsafe.Pointer(self))\n", typ.gofmt())
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

	g.genTypeTPCall(typ)

	g.genTypeMethods(typ)

}

func (g *goGen) genTypeTPCall(typ Type) {
	if !typ.isSignature() {
		return
	}

	sig := typ.GoType().Underlying().(*types.Signature)
	if sig.Recv() != nil {
		// don't generate tp_call for methods.
		return
	}

	// support for __call__
	g.Printf("//export cgo_func_%[1]s_call\n", typ.ID())
	g.Printf("func cgo_func_%[1]s_call(self %[2]s", typ.ID(), typ.CGoTypeName())
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
	g.Printf("(*(*%[1]s)(unsafe.Pointer(self)))(", typ.gofmt())
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

func (g *goGen) genTypeMethods(typ Type) {
	if !typ.isNamed() {
		return
	}

	tn := typ.GoType().(*types.Named)
	for imeth := 0; imeth < tn.NumMethods(); imeth++ {
		m := tn.Method(imeth)
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
			typ.CGoName(),
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
		if typ.isBasic() {
			g.Printf("(*%s)(unsafe.Pointer(&self)).%s(",
				typ.gofmt(),
				msym.goname,
			)
		} else {
			g.Printf("(*%s)(unsafe.Pointer(self)).%s(",
				typ.gofmt(),
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
