// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
)

func (g *pybindGen) genVarGetter(v Var) {
	pkgname := g.pkg.Name()
	cgoFn := v.Name() // plain name is the getter
	qFn := "_" + pkgname + "." + cgoFn
	qVn := pkgname + "." + v.Name()

	g.pywrap.Printf("def %s():\n", cgoFn)
	g.pywrap.Indent()
	g.pywrap.Printf("%s\n%s Gets Go Variable: %s\n%s\n%s\n", `"""`, cgoFn, qVn, v.doc, `"""`)
	if v.sym.hasHandle() {
		g.pywrap.Printf("return %s(handle=%s())\n", v.sym.pyname, qFn)
	} else {
		g.pywrap.Printf("return %s()\n", qFn)
	}
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", cgoFn)
	g.gofile.Printf("func %s() %s {\n", cgoFn, v.sym.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("return ")
	if v.sym.go2py != "" {
		if v.sym.hasHandle() && !v.sym.isPtrOrIface() {
			g.gofile.Printf("%s(&%s)", v.sym.go2py, qVn)
		} else {
			g.gofile.Printf("%s(%s)", v.sym.go2py, qVn)
		}
	} else {
		g.gofile.Printf("%s", qVn)
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', retval('%s'), [])\n", cgoFn, v.sym.cpyname)
}

func (g *pybindGen) genVarSetter(v Var) {
	pkgname := g.pkg.Name()
	cgoFn := fmt.Sprintf("Set%s", v.Name())
	qFn := "_" + pkgname + "." + cgoFn
	qVn := pkgname + "." + v.Name()

	g.pywrap.Printf("def %s(value):\n", cgoFn)
	g.pywrap.Indent()
	g.pywrap.Printf("%s\n%s Sets Go Variable: %s\n%s\n%s\n", `"""`, cgoFn, qVn, v.doc, `"""`)
	g.pywrap.Printf("if isinstance(value, GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("%s(value.handle)\n", qFn)
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("%s(value)\n", qFn)
	g.pywrap.Outdent()
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", cgoFn)
	g.gofile.Printf("func %s(val %s) {\n", cgoFn, v.sym.cgoname)
	g.gofile.Indent()
	if v.sym.go2py != "" {
		if v.sym.hasHandle() && !v.sym.isPtrOrIface() {
			g.gofile.Printf("%s = *%s(val)", qVn, v.sym.py2go)
		} else {
			g.gofile.Printf("%s = %s(val)", qVn, v.sym.py2go)
		}
	} else {
		g.gofile.Printf("%s = val", qVn)
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', None, [param('%s', 'val')])\n", cgoFn, v.sym.cpyname)
}

func (g *pybindGen) genConstValue(c Const) {
	// constants go directly into wrapper as-is
	g.pywrap.Printf("%s = %s\n", c.GoName(), c.obj.Val().ExactString())
}
