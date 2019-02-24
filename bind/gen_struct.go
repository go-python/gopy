// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
)

func (g *pyGen) genStruct(s *Struct) {
	strNm := s.obj.Name()
	g.pywrap.Printf(`
# Python type for struct %[3]s
class %[1]s(go.GoClass):
	""%[2]q""
`,
		strNm,
		s.Doc(),
		s.GoName(),
	)
	g.pywrap.Indent()
	g.genStructInit(s)
	g.genStructMembers(s)
	g.genStructMethods(s)
	g.pywrap.Outdent()
}

func (g *pyGen) genStructInit(s *Struct) {
	pkgname := g.outname
	qNm := s.GoName()

	numFields := s.Struct().NumFields()

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
	g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], go.GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = args[0].handle\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = _%s.%s_CTor()\n", pkgname, s.ID())

	for i := 0; i < numFields; i++ {
		f := s.Struct().Field(i)
		if !f.Exported() || f.Embedded() {
			continue
		}
		ftyp := current.symtype(f.Type())
		if ftyp == nil || ftyp.isSignature() || (ftyp.isPointer() && ftyp.isBasic()) {
			continue
		}
		g.pywrap.Printf("if  %[1]d < len(args):\n", i)
		g.pywrap.Indent()
		g.pywrap.Printf("self.%s = args[%d]\n", f.Name(), i)
		g.pywrap.Outdent()
		g.pywrap.Printf("if %[1]q in kwargs:\n", f.Name())
		g.pywrap.Indent()
		g.pywrap.Printf("self.%[1]s = kwargs[%[1]q]\n", f.Name())
		g.pywrap.Outdent()
	}
	g.pywrap.Outdent()
	g.pywrap.Outdent()

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
	ctNm := s.ID() + "_CTor"
	g.gofile.Printf("\n// --- wrapping struct: %v ---\n", qNm)
	g.gofile.Printf("//export %s\n", ctNm)
	g.gofile.Printf("func %s() CGoHandle {\n", ctNm)
	g.gofile.Indent()
	g.gofile.Printf("return CGoHandle(handleFmPtr_%s(&%s{}))\n", s.ID(), qNm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")

	g.pybuild.Printf("mod.add_function('%s', retval('%s'), [])\n", ctNm, PyHandle)

}

func (g *pyGen) genStructMembers(s *Struct) {
	typ := s.Struct()
	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() || f.Embedded() {
			continue
		}
		ftyp := current.symtype(f.Type())
		if ftyp == nil || ftyp.isSignature() || (ftyp.isPointer() && ftyp.isBasic()) {
			continue
		}
		g.genStructMemberGetter(s, i, f)
		g.genStructMemberSetter(s, i, f)
	}
}

func (g *pyGen) genStructMemberGetter(s *Struct, i int, f types.Object) {
	pkgname := g.outname
	ft := f.Type()
	ret := current.symtype(ft)
	if ret == nil {
		return
	}

	cgoFn := fmt.Sprintf("%s_%s_Get", s.ID(), f.Name())

	g.pywrap.Printf("@property\n")
	g.pywrap.Printf("def %[1]s(self):\n", f.Name())
	g.pywrap.Indent()
	if ret.hasHandle() {
		cvnm := ret.pyPkgId(g.pkg.pkg)
		g.pywrap.Printf("return %s(handle=_%s.%s(self.handle))\n", cvnm, pkgname, cgoFn)
	} else {
		g.pywrap.Printf("return _%s.%s(self.handle)\n", pkgname, cgoFn)
	}
	g.pywrap.Outdent()

	g.gofile.Printf("//export %s\n", cgoFn)
	g.gofile.Printf("func %s(handle CGoHandle) %s {\n", cgoFn, ret.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("op := ptrFmHandle_%s(handle)\nreturn ", s.ID())
	if ret.go2py != "" {
		if ret.hasHandle() && !ret.isPtrOrIface() {
			g.gofile.Printf("%s(&op.%s)%s", ret.go2py, f.Name(), ret.go2pyParenEx)
		} else {
			g.gofile.Printf("%s(op.%s)%s", ret.go2py, f.Name(), ret.go2pyParenEx)
		}
	} else {
		g.gofile.Printf("op.%s", f.Name())
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', retval('%s'), [param('%s', 'handle')])\n", cgoFn, ret.cpyname, PyHandle)
}

func (g *pyGen) genStructMemberSetter(s *Struct, i int, f types.Object) {
	pkgname := g.outname
	ft := f.Type()
	ret := current.symtype(ft)
	if ret == nil {
		return
	}

	cgoFn := fmt.Sprintf("%s_%s_Set", s.ID(), f.Name())

	g.pywrap.Printf("@%s.setter\n", f.Name())
	g.pywrap.Printf("def %[1]s(self, value):\n", f.Name())
	g.pywrap.Indent()
	g.pywrap.Printf("if isinstance(value, go.GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("_%s.%s(self.handle, value.handle)\n", pkgname, cgoFn)
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("_%s.%s(self.handle, value)\n", pkgname, cgoFn)
	g.pywrap.Outdent()
	g.pywrap.Outdent()

	g.gofile.Printf("//export %s\n", cgoFn)
	g.gofile.Printf("func %s(handle CGoHandle, val %s) {\n", cgoFn, ret.cgoname)
	g.gofile.Indent()
	g.gofile.Printf("op := ptrFmHandle_%s(handle)\n", s.ID())
	if ret.py2go != "" {
		g.gofile.Printf("op.%s = %s(val)%s", f.Name(), ret.py2go, ret.py2goParenEx)
	} else {
		g.gofile.Printf("op.%s = val", f.Name())
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pybuild.Printf("mod.add_function('%s', None, [param('%s', 'handle'), param('%s', 'val')])\n", cgoFn, PyHandle, ret.cpyname)
}

func (g *pyGen) genStructMethods(s *Struct) {
	for _, m := range s.meths {
		g.genMethod(s, &m)
	}
}

//////////////////////////////////////////////////////////////////////////
// Interface

func (g *pyGen) genInterface(ifc *Interface) {
	strNm := ifc.obj.Name()
	g.pywrap.Printf(`
# Python type for interface %[3]s
class %[1]s(go.GoClass):
	""%[2]q""
`,
		strNm,
		ifc.Doc(),
		ifc.GoName(),
	)
	g.pywrap.Indent()
	g.genIfaceInit(ifc)
	g.genIfaceMethods(ifc)
	g.pywrap.Outdent()
}

func (g *pyGen) genIfaceInit(ifc *Interface) {
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
	g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], go.GoClass):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = args[0].handle\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = 0\n")
	g.pywrap.Outdent()
	g.pywrap.Outdent()

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

func (g *pyGen) genIfaceMethods(ifc *Interface) {
	for _, m := range ifc.meths {
		g.genIfcMethod(ifc, &m)
	}
}
