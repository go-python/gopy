// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

func (g *pyGen) genConst(c *Const) {
	if isPyCompatVar(c.sym) != nil {
		return
	}
	if c.sym.isSignature() {
		return
	}
	g.genConstValue(c)
}

func (g *pyGen) genVar(v *Var) {
	if isPyCompatVar(v.sym) != nil {
		return
	}
	if v.sym.isSignature() {
		return
	}
	g.genVarGetter(v)
	if !v.sym.isArray() {
		g.genVarSetter(v)
	}
}

func (g *pyGen) genVarGetter(v *Var) {
	gopkg := g.pkg.Name()
	pkgname := g.cfg.Name
	cgoFn := v.Name() // plain name is the getter
	if g.cfg.RenameCase {
		cgoFn = toSnakeCase(cgoFn)
	}
	qCgoFn := gopkg + "_" + cgoFn
	qFn := "_" + pkgname + "." + qCgoFn
	qVn := gopkg + "." + v.Name()

	g.pywrap.Printf("def %s():\n", cgoFn)
	g.pywrap.Indent()
	g.pywrap.Printf("%s\n%s Gets Go Variable: %s\n%s\n%s\n", `"""`, cgoFn, qVn, v.doc, `"""`)
	if v.sym.hasHandle() {
		cvnm := v.sym.pyPkgId(g.pkg.pkg)
		g.pywrap.Printf("return %s(handle=%s())\n", cvnm, qFn)
	} else {
		g.pywrap.Printf("return %s()\n", qFn)
	}
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", qCgoFn)
	g.gofile.Printf("func %s() %s {\n", qCgoFn, v.sym.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("return ")
	if v.sym.go2py != "" {
		if v.sym.hasHandle() && !v.sym.isPtrOrIface() {
			g.gofile.Printf("%s(&%s)%s", v.sym.go2py, qVn, v.sym.go2pyParenEx)
		} else {
			g.gofile.Printf("%s(%s)%s", v.sym.go2py, qVn, v.sym.go2pyParenEx)
		}
	} else {
		g.gofile.Printf("%s", qVn)
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', retval('%s'), [])\n", qCgoFn, v.sym.cpyname)
}

func (g *pyGen) genVarSetter(v *Var) {
	gopkg := g.pkg.Name()
	pkgname := g.cfg.Name
	cgoFn := fmt.Sprintf("Set_%s", v.Name())
	if g.cfg.RenameCase {
		cgoFn = toSnakeCase(cgoFn)
	}
	qCgoFn := gopkg + "_" + cgoFn
	qFn := "_" + pkgname + "." + qCgoFn
	qVn := gopkg + "." + v.Name()

	g.pywrap.Printf("def %s(value):\n", cgoFn)
	g.pywrap.Indent()
	g.pywrap.Printf("%s\n%s Sets Go Variable: %s\n%s\n%s\n", `"""`, cgoFn, qVn, v.doc, `"""`)
	g.pywrap.Printf("if isinstance(value, go.GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("%s(value.handle)\n", qFn)
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("%s(value)\n", qFn)
	g.pywrap.Outdent()
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", qCgoFn)
	g.gofile.Printf("func %s(val %s) {\n", qCgoFn, v.sym.cgoname)
	g.gofile.Indent()
	if v.sym.py2go != "" {
		g.gofile.Printf("%s = %s(val)%s", qVn, v.sym.py2go, v.sym.py2goParenEx)
	} else {
		g.gofile.Printf("%s = val", qVn)
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', None, [param('%s', 'val')])\n", qCgoFn, v.sym.cpyname)
}

func (g *pyGen) genConstValue(c *Const) {
	// constants go directly into wrapper as-is
	val := c.val
	switch val {
	case "true":
		val = "True"
	case "false":
		val = "False"
	}
	g.pywrap.Printf("%s = %s\n", c.GoName(), val)
}

func (g *pyGen) genEnum(e *Enum) {
	g.pywrap.Printf("class %s(Enum):\n", e.typ.Obj().Name())
	g.pywrap.Indent()
	doc := e.Doc()
	if doc != "" {
		lns := strings.Split(doc, "\n")
		g.pywrap.Printf(`"""`)
		g.pywrap.Printf("\n")
		for _, l := range lns {
			g.pywrap.Printf("%s\n", l)
		}
		g.pywrap.Printf(`"""`)
		g.pywrap.Printf("\n")
	}
	e.SortConsts()
	for _, c := range e.items {
		g.genConstValue(c)
	}
	g.pywrap.Outdent()

	// Go has each const value globally available within a given package
	// so to keep the code consistent, we redundantly generate the consts
	// again here.  The Enum organization however is critical for organizing
	// the values under the type (making them accessible programmatically)
	g.pywrap.Printf("\n")
	for _, c := range e.items {
		g.genConstValue(c)
	}
	g.pywrap.Printf("\n")
}
