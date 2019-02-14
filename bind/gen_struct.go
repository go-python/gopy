// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
)

func (g *pybindGen) genStruct(s Struct) {
	pkgname := s.Package().Name()
	g.pywrap.Printf(`
# Python type for struct %[1]s.%[2]s
class %[2]s(GoClass):
	""%[3]q""
`,
		pkgname,
		s.GoName(),
		s.Doc(),
	)
	g.pywrap.Indent()
	g.genStructInit(s)
	g.genStructMembers(s)
	g.genStructMethods(s)
	g.pywrap.Outdent()
}

func (g *pybindGen) genStructInit(s Struct) {
	// pkg := s.Package()
	pkgname := s.Package().Name()
	qNm := pkgname + "." + s.GoName()

	numFields := s.Struct().NumFields()
	numPublic := numFields
	for i := 0; i < s.Struct().NumFields(); i++ {
		f := s.Struct().Field(i)
		if !f.Exported() {
			numPublic--
			continue
		}
	}

	g.pywrap.Printf("def __init__(self, *args, **kwargs):\n")
	g.pywrap.Indent()
	g.pywrap.Printf(`"""
handle=A Go-side object is always initialized with an explicit handle=arg
otherwise parameters can be unnamed in order of field names or named fields
in which case a new Go object is constructed first
"""
`)
	g.pywrap.Printf("if len(kwargs) == 1 and 'handle' in kwargs:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = kwargs['handle']\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = _%s_CTor()\n", qNm)

	for i := 0; i < numFields; i++ {
		field := s.Struct().Field(i)
		if !field.Exported() {
			continue
		}
		g.pywrap.Printf("if  %[1]d < len(args):\n", i)
		g.pywrap.Indent()
		g.pywrap.Printf("self.%s = args[%d]\n", field.Name(), i)
		g.pywrap.Outdent()
		g.pywrap.Printf("if %[1]q in kwargs:\n", field.Name())
		g.pywrap.Indent()
		g.pywrap.Printf("self.%[1]s = kwargs[%[1]q]\n", field.Name())
		g.pywrap.Outdent()
	}
	g.pywrap.Outdent()
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	for _, m := range s.meths {
		if m.GoName() == "String" {
			g.pywrap.Printf("def __str__(self):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("return self.String()\n")
			g.pywrap.Outdent()
			g.pywrap.Printf("\n")
		}
	}

	// go ctor
	ctNm := s.GoName() + "_CTor"
	g.gofile.Printf("\n// --- wrapping struct: %v ---\n", qNm)
	g.gofile.Printf("//export %s\n", ctNm)
	g.gofile.Printf("func %s() CGoHandle {\n", ctNm)
	g.gofile.Indent()
	g.gofile.Printf("return handleFmPtr_%[1]s(&%[2]s{})\n", s.GoName(), qNm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")

	g.pybuild.Printf("mod.add_function('%s', retval('char*'), [])\n", ctNm)

}

func (g *pybindGen) genStructMembers(s Struct) {
	//pkgname := s.Package().Name()
	typ := s.Struct()
	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() {
			continue
		}
		g.genStructMemberGetter(s, i, f)
		g.genStructMemberSetter(s, i, f)
	}
}

func (g *pybindGen) genStructMemberGetter(s Struct, i int, f types.Object) {
	pkgname := s.Package().Name()
	ft := f.Type()
	ret := g.pkg.syms.symtype(ft)

	cgoFn := fmt.Sprintf("%s_%s_Get", s.GoName(), f.Name())

	g.pywrap.Printf("@property\n")
	g.pywrap.Printf("def %[1]s(self):\n", f.Name())
	g.pywrap.Indent()
	if ret.hasHandle() {
		g.pywrap.Printf("return %s(handle=_%s.%s(self.handle))\n", ret.nonPointerName(), pkgname, cgoFn)
	} else {
		g.pywrap.Printf("return _%s.%s(self.handle)\n", pkgname, cgoFn)
	}
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", cgoFn)
	g.gofile.Printf("func %s(handle CGoHandle) %s {\n", cgoFn, ret.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("op := ptrFmHandle_%s(handle)\nreturn ", s.GoName())
	if ret.go2py != "" {
		if ret.hasHandle() && !ret.isPtrOrIface() {
			g.gofile.Printf("%s(&op.%s)", ret.go2py, f.Name())
		} else {
			g.gofile.Printf("%s(op.%s)", ret.go2py, f.Name())
		}
	} else {
		g.gofile.Printf("op.%s", f.Name())
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', retval('%s'), [param('char*', 'handle')])\n", cgoFn, ret.cpyname)
}

func (g *pybindGen) genStructMemberSetter(s Struct, i int, f types.Object) {
	pkgname := s.Package().Name()
	ft := f.Type()
	ret := g.pkg.syms.symtype(ft)

	cgoFn := fmt.Sprintf("%s_%s_Set", s.GoName(), f.Name())

	g.pywrap.Printf("@%s.setter\n", f.Name())
	g.pywrap.Printf("def %[1]s(self, value):\n", f.Name())
	g.pywrap.Indent()
	g.pywrap.Printf("if isinstance(value, GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("_%s.%s(self.handle, value.handle)\n", pkgname, cgoFn)
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("_%s.%s(self.handle, value)\n", pkgname, cgoFn)
	g.pywrap.Outdent()
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", cgoFn)
	g.gofile.Printf("func %s(handle CGoHandle, val %s) {\n", cgoFn, ret.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("op := ptrFmHandle_%s(handle)\n", s.GoName())
	if ret.go2py != "" {
		if ret.hasHandle() && !ret.isPtrOrIface() {
			g.gofile.Printf("op.%s = *%s(val)", f.Name(), ret.py2go)
		} else {
			g.gofile.Printf("op.%s = %s(val)", f.Name(), ret.py2go)
		}
	} else {
		g.gofile.Printf("op.%s = val", f.Name())
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', None, [param('char*', 'handle'), param('%s', 'val')])\n", cgoFn, ret.cpyname)
}

func (g *pybindGen) genStructMethods(s Struct) {
	for _, m := range s.meths {
		g.genMethod(s, m)
	}
}

//////////////////////////////////////////////////////////////////////////
// Interface

func (g *pybindGen) genInterface(ifc Interface) {
	pkgname := ifc.Package().Name()
	g.pywrap.Printf(`
# Python type for interface %[1]s.%[2]s
class %[2]s(GoClass):
	""%[3]q""
`,
		pkgname,
		ifc.GoName(),
		ifc.Doc(),
	)
	g.pywrap.Indent()
	g.genIfaceInit(ifc)
	g.genIfaceMethods(ifc)
	g.pywrap.Outdent()
}

func (g *pybindGen) genIfaceInit(ifc Interface) {
	// pkgname := ifc.Package().Name()

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
	g.pywrap.Printf("self.handle = args[0]\n")
	g.pywrap.Outdent()
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	for _, m := range ifc.meths {
		if m.GoName() == "String" {
			g.pywrap.Printf("def __str__(self):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("return self.String()\n")
			g.pywrap.Outdent()
			g.pywrap.Printf("\n")
		}
	}
}

func (g *pybindGen) genIfaceMethods(ifc Interface) {
	for _, m := range ifc.meths {
		g.genIfcMethod(ifc, m)
	}
}
