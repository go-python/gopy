// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

func (g *pybindGen) genType(sym *symbol) {
	if !sym.isType() {
		return
	}
	if sym.isBasic() && !sym.isNamed() {
		return
	}

	if sym.isNamedBasic() {
		// todo: could have methods!
		return
	}

	// todo: not handling yet:
	if sym.isSignature() {
		return
	}

	if sym.isPointer() || sym.isInterface() {
		g.genTypeHandlePtr(sym)
	} else {
		g.genTypeHandle(sym)
	}

	if sym.isSlice() {
		g.genSlice(sym)
	}
}

func (g *pybindGen) genTypeHandlePtr(sym *symbol) {
	gonm := sym.gofmt()
	g.gofile.Printf("\n// Converters for pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", sym.py2go, gonm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarHand.VarFmHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return p.(%[1]s)\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func %s(p interface{}) CGoHandle {\n", sym.go2py)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(gopyh.VarHand.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

func (g *pybindGen) genTypeHandle(sym *symbol) {
	gonm := sym.gofmt()
	ptrnm := "*" + gonm
	py2go := sym.py2go
	if py2go[0] == '*' {
		py2go = py2go[1:]
	}
	g.gofile.Printf("\n// Converters for non-pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", py2go, ptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarHand.VarFmHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return p.(%[1]s)\n", ptrnm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func %s(p interface{}) CGoHandle {\n", sym.go2py)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(gopyh.VarHand.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}
