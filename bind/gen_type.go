// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

// extTypes = these are types external to any targeted packages
// pyWrapOnly = only generate python wrapper code, not go code
func (g *pyGen) genType(sym *symbol, extTypes, pyWrapOnly bool) {
	if !sym.isType() {
		return
	}
	if sym.isBasic() && !sym.isNamed() {
		return
	}

	if sym.isNamedBasic() {
		// TODO: could have methods!
		return
	}

	// TODO: not handling yet:
	if sym.isSignature() {
		return
	}

	if !pyWrapOnly {
		switch {
		case sym.isPointer() || sym.isInterface():
			g.genTypeHandlePtr(sym)
		case sym.isSlice() || sym.isMap() || sym.isArray():
			g.genTypeHandleImplPtr(sym)
		default:
			g.genTypeHandle(sym)
		}
	}

	if extTypes {
		if sym.isSlice() || sym.isArray() {
			g.genSlice(sym, extTypes, pyWrapOnly, nil)
		} else if sym.isMap() {
			g.genMap(sym, extTypes, pyWrapOnly, nil)
		} else if sym.isInterface() || sym.isStruct() {
			if pyWrapOnly {
				g.genExtClass(sym)
			}
		}
	} else {
		if g.pkg == goPackage || !sym.isNamed() { // only named types are generated separately
			if sym.isSlice() {
				g.genSlice(sym, extTypes, pyWrapOnly, nil)
			} else if sym.isMap() {
				g.genMap(sym, extTypes, pyWrapOnly, nil)
			}
		}
		if sym.isArray() {
			g.genSlice(sym, extTypes, pyWrapOnly, nil)
		}
	}
}

func (g *pyGen) genTypeHandlePtr(sym *symbol) {
	if sym.goname == "interface{}" {
		return
	}
	gonm := sym.goname
	g.gofile.Printf("\n// Converters for pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", sym.py2go, gonm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarFromHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	if sym.isStruct() {
		g.gofile.Printf("return gopyh.Embed(p, reflect.TypeOf(%s{})).(%s)\n", nonPtrName(gonm), gonm)
	} else {
		g.gofile.Printf("return p.(%s)\n", gonm)
	}
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func %s(p interface{})%s CGoHandle {\n", sym.go2py, sym.go2pyParenEx)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(gopyh.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

// implicit pointer types: slice, map, array
func (g *pyGen) genTypeHandleImplPtr(sym *symbol) {
	gonm := sym.goname
	ptrnm := gonm
	nptrnm := gonm
	if ptrnm[0] != '*' {
		ptrnm = "*" + ptrnm
	} else {
		nptrnm = gonm[1:]
	}
	g.gofile.Printf("\n// Converters for implicit pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func ptrFromHandle_%s(h CGoHandle) %s {\n", sym.id, ptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarFromHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return p.(%s)\n", ptrnm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func deptrFromHandle_%s(h CGoHandle) %s {\n", sym.id, nptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := ptrFromHandle_%s(h)\n", sym.id)
	if !sym.isArray() {
		g.gofile.Printf("if p == nil {\n")
		g.gofile.Indent()
		g.gofile.Printf("return nil\n")
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
	}
	g.gofile.Printf("return *p\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func %s(p interface{})%s CGoHandle {\n", sym.go2py, sym.go2pyParenEx)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(gopyh.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

func nonPtrName(nm string) string {
	if nm[0] == '*' {
		return nm[1:]
	}
	return nm
}

func (g *pyGen) genTypeHandle(sym *symbol) {
	gonm := sym.goname
	ptrnm := gonm
	if ptrnm[0] != '*' {
		ptrnm = "*" + ptrnm
	}
	py2go := nonPtrName(sym.py2go)
	g.gofile.Printf("\n// Converters for non-pointer handles for type: %s\n", gonm)
	g.gofile.Printf("func %s(h CGoHandle) %s {\n", py2go, ptrnm)
	g.gofile.Indent()
	g.gofile.Printf("p := gopyh.VarFromHandle((gopyh.CGoHandle)(h), %[1]q)\n", gonm)
	g.gofile.Printf("if p == nil {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	if sym.isStruct() {
		g.gofile.Printf("return gopyh.Embed(p, reflect.TypeOf(%s{})).(%s)\n", nonPtrName(gonm), ptrnm)
	} else {
		g.gofile.Printf("return p.(%s)\n", ptrnm)
	}
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("func %s(p interface{})%s CGoHandle {\n", sym.go2py, sym.go2pyParenEx)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(gopyh.Register(\"%s\", p))\n", gonm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}

// genExtClass generates minimal python wrappers for external classes (struct, interface, etc)
func (g *pyGen) genExtClass(sym *symbol) {
	pkgname := sym.gopkg.Name()
	// note: all external wrapper classes are defined in base go.py module, so we exclude go.
	g.pywrap.Printf(`
# Python type for %[4]s
class %[2]s(GoClass):
	""%[3]q""
`,
		pkgname,
		sym.id,
		sym.doc,
		sym.goname,
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
	g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
	g.pywrap.Outdent()
	g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = args[0].handle\n")
	g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
	g.pywrap.Outdent()
	g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], int):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = args[0]\n")
	g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = 0\n")
	g.pywrap.Outdent()
	g.pywrap.Outdent()

	g.pywrap.Printf("def __del__(self):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("_%s.DecRef(self.handle)\n", g.pypkgname)
	g.pywrap.Outdent()

	g.pywrap.Printf("\n")
	g.pywrap.Outdent()

}
