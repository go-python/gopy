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
	if sym.isArray() || sym.isSlice() {
		return
	}

	if sym.isPointer() {
		g.genTypeHandlePtr(sym)
	} else {
		g.genTypeHandle(sym)
	}
}

func (g *pybindGen) genTypeHandlePtr(sym *symbol) {
	npnm := sym.nonPointerName() + "_Ptr"
	g.gofile.Printf("\n// Converters for pointer handles for type: %s\n", sym.gofmt())
	g.gofile.Printf("func ptrFmHandle_%s(h *C.char) %s {\n", npnm, sym.gofmt())
	g.gofile.Indent()
	g.gofile.Printf("p := varHand.varFmHandle(h, %[1]q)\n", sym.gofmt())
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return p.(%[1]s)\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func handleFmPtr_%s(p interface{}) *C.char {\n", npnm)
	g.gofile.Indent()
	g.gofile.Printf("return varHand.register(\"%s\", p)\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

func (g *pybindGen) genTypeHandle(sym *symbol) {
	npnm := sym.nonPointerName()
	ptrnm := "*" + sym.gofmt()
	g.gofile.Printf("\n// Converters for pointer handles for type: %s\n", sym.gofmt())
	g.gofile.Printf("func ptrFmHandle_%s(h *C.char) %s {\n", npnm, ptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := varHand.varFmHandle(h, %[1]q)\n", sym.gofmt())
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return p.(%[1]s)\n", ptrnm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func handleFmPtr_%s(p interface{}) *C.char {\n", npnm)
	g.gofile.Indent()
	g.gofile.Printf("return varHand.register(\"%s\", p)\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}
