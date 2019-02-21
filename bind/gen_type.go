// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

func (g *pyGen) genType(sym *symbol) {
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
	} else if sym.isInterface() || sym.isStruct() {
		if g.pkg.pkg != sym.gopkg {
			g.genExtClass(sym)
		}
	}
}

func (g *pyGen) genTypeHandlePtr(sym *symbol) {
	gonm := sym.gofmt()
	g.gofile.Printf("\n// Converters for pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", sym.py2go, gonm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarFmHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
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
	g.gofile.Printf("return CGoHandle(gopyh.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

func (g *pyGen) genTypeHandle(sym *symbol) {
	gonm := sym.gofmt()
	ptrnm := gonm
	if ptrnm[:1] != "*" {
		ptrnm = "*" + ptrnm
	}
	py2go := sym.py2go
	if py2go[0] == '*' {
		py2go = py2go[1:]
	}
	g.gofile.Printf("\n// Converters for non-pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", py2go, ptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarFmHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
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
	g.gofile.Printf("return CGoHandle(gopyh.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

// genExtClass generates minimal python wrappers for external classes (struct, interface, etc)
func (g *pyGen) genExtClass(sym *symbol) {
	pkgname := sym.gopkg.Name()
	g.pywrap.Printf(`
# Python type for interface %[1]s.%[2]s
class %[2]s(GoClass):
	""%[3]q""
`,
		pkgname,
		sym.id,
		sym.doc,
	)
	g.pywrap.Indent()
	g.pywrap.Printf("def __init__(self, *args, **kwargs):\n")
	g.pywrap.Indent()
	g.pywrap.Printf(`"""
handle=A Go-side object is always initialized with an explicit handle=arg
"""
`)
	g.pywrap.Printf("if len(kwargs) == 1 and 'handle' in kwargs:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = kwargs['handle']\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = 0\n")
	g.pywrap.Outdent()
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")
	g.pywrap.Outdent()
}
