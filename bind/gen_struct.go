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
class %[2]s:
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
	// g.genStructTPRepr(s)
	// g.genStructTPStr(s)
	g.pywrap.Outdent()
}

func (g *pybindGen) genStructInit(s Struct) {
	// pkg := s.Package()
	// pkgname := s.Package().Name()
	numFields := s.Struct().NumFields()
	numPublic := numFields
	for i := 0; i < s.Struct().NumFields(); i++ {
		f := s.Struct().Field(i)
		if !f.Exported() {
			numPublic--
			continue
		}
	}

	g.pywrap.Printf("def __init__(self, handle):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = handle\n")
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

	cgo_fgetname := fmt.Sprintf("%s_%s_Get", s.GoName(), f.Name())

	g.pywrap.Printf("@property\n")
	g.pywrap.Printf("def %[1]s(self):\n", f.Name())
	g.pywrap.Indent()
	g.pywrap.Printf("return _%s.%s(self.handle)\n", pkgname, cgo_fgetname)
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s\n", cgo_fgetname)
	g.gofile.Printf("func %s(handle *C.char) %s {\n", cgo_fgetname, ret.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("op := ptrFmHandle_%s(handle)\nreturn ", s.GoName())
	if ret.c2py != "" {
		g.gofile.Printf("%s(op.%s)", ret.c2py, f.Name())
	} else {
		g.gofile.Printf("op.%s", f.Name())
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', retval('%s'), [param('char*', 'handle')])\n", cgo_fgetname, ret.cpyname)
}

func (g *pybindGen) genStructMemberSetter(s Struct, i int, f types.Object) {
	pkgname := s.Package().Name()
	// ft := f.Type()
	cgo_fsetname := fmt.Sprintf("%s_%s_Set", s.GoName(), f.Name())

	g.pywrap.Printf("@%[1]s.setter\n", f.Name())
	g.pywrap.Printf("def %[1]s(self, value):\n", f.Name())
	g.pywrap.Indent()
	g.pywrap.Printf("_%s.%s(self.handle, value)\n", pkgname, cgo_fsetname)
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")
}

func (g *pybindGen) genStructMethods(s Struct) {
	for _, m := range s.meths {
		g.genMethod(s, m)
	}
}
