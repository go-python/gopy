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

	if sym.isPointer() {
		g.genTypeHandlePtr(sym)
	} else {
		g.genTypeHandle(sym)
	}

	if sym.isSlice() {
		g.genSlice(sym)
	}
}

func (g *pybindGen) genTypeHandlePtr(sym *symbol) {
	g.gofile.Printf("\n// Converters for pointer handles for type: %s\n", sym.gofmt())
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", sym.py2go, sym.gofmt())
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarHand.VarFmHandle((gopyh.CGoHandle)(h), %[1]q)\n", sym.gofmt())
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return p.(%[1]s)\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func %s(p interface{}) CGoHandle {\n", sym.go2py)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(gopyh.VarHand.Register(\"%s\", p))\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

func (g *pybindGen) genTypeHandle(sym *symbol) {
	ptrnm := "*" + sym.gofmt()
	g.gofile.Printf("\n// Converters for non-pointer handles for type: %s\n", sym.gofmt())
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", sym.py2go, ptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarHand.VarFmHandle((gopyh.CGoHandle)(h), %[1]q)\n", sym.gofmt())
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
	g.gofile.Printf("return CGoHandle(gopyh.VarHand.Register(\"%s\", p))\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}
