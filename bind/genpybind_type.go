// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

func (g *pybindGen) genType(sym *symbol) {
	if !sym.isType() {
		return
	}
	// if sym.isStruct() {
	// 	return
	// }
	if sym.isBasic() && !sym.isNamed() {
		return
	}
	// if sym.isPointer() {
	// 	return
	// }

	if sym.isPointer() {
		g.genTypePointer(sym)
	}
}

func (g *pybindGen) genTypePointer(sym *symbol) {
	npnm := sym.goname[1:] // non-pointer name
	g.gofile.Printf("\n// Converters for pointer handles for type: %v\n", sym.gofmt())
	g.gofile.Printf("func ptrFmHandle_%v(h *C.char) %v {\n", npnm, sym.gofmt())
	g.gofile.Indent()
	g.gofile.Printf("return varHand.varFmHandle(h, \"%[1]s\").(%[1]s)\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func handleFmPtr_%v(p interface{}) *C.char {\n", npnm)
	g.gofile.Indent()
	g.gofile.Printf("return varHand.register(\"%v\", p)\n", sym.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}
