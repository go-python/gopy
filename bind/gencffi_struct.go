// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
)

func (g *cffiGen) genStruct(s Struct) {
	pkgname := s.Package().Name()
	g.wrapper.Printf(`
# Python type for struct %[1]s.%[2]s
class %[2]s(object):
    ""%[3]q""
`,
		pkgname,
		s.GoName(),
		s.Doc(),
	)
	g.wrapper.Indent()
	g.genStructInit(s)
	g.genStructMembers(s)
	g.genStructMethods(s)
	g.genStructTPRepr(s)
	g.genStructTPStr(s)
	g.wrapper.Outdent()
}

// FIXME: Conversion function should be improved to more memory efficiency way.
func (g *cffiGen) genStructConversion(s Struct) {
	g.wrapper.Printf("@staticmethod\n")
	g.wrapper.Printf("def cffi_cgopy_cnv_py2c_%[1]s_%[2]s(o):\n", s.Package().Name(), s.GoName())
	g.wrapper.Indent()
	g.wrapper.Printf("return o.cgopy\n\n")
	g.wrapper.Outdent()

	g.wrapper.Printf("@staticmethod\n")
	g.wrapper.Printf("def cffi_cgopy_cnv_c2py_%[1]s_%[2]s(c):\n", s.Package().Name(), s.GoName())
	g.wrapper.Indent()
	g.wrapper.Printf("o = %[1]s()\n", s.GoName())
	g.wrapper.Printf("o.cgopy = ffi.gc(c, _cffi_helper.lib.cgopy_decref)\n")
	g.wrapper.Printf("return o\n\n")
	g.wrapper.Outdent()
}

func (g *cffiGen) genStructInit(s Struct) {
	pkg := s.Package()
	pkgname := s.Package().Name()
	numFields := s.Struct().NumFields()
	numPublic := numFields
	for i := 0; i < s.Struct().NumFields(); i++ {
		f := s.Struct().Field(i)
		if !f.Exported() {
			numPublic--
			continue
		}
	}

	g.wrapper.Printf("def __init__(self, *args, **kwargs):\n")
	g.wrapper.Indent()
	if numPublic > 0 {
		g.wrapper.Printf("nkwds = len(kwargs)\n")
		g.wrapper.Printf("nargs = len(args)\n")
		g.wrapper.Printf("if nkwds + nargs > %[1]d:\n", numPublic)
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%s.__init__ takes at most %d argument(s)')\n",
			s.GoName(),
			numPublic,
		)
		g.wrapper.Outdent()
	}
	g.wrapper.Printf("cgopy = _cffi_helper.lib.cgo_func_%[1]s_%[2]s_new()\n", pkgname, s.GoName())
	g.wrapper.Printf("if cgopy == ffi.NULL:\n")
	g.wrapper.Indent()
	g.wrapper.Printf("raise MemoryError('gopy: could not allocate %[1]s.')\n", s.GoName())
	g.wrapper.Outdent()
	g.wrapper.Printf("self.cgopy = ffi.gc(cgopy, _cffi_helper.lib.cgopy_decref)\n")
	g.wrapper.Printf("\n")

	for i := 0; i < numFields; i++ {
		field := s.Struct().Field(i)
		if !field.Exported() {
			continue
		}

		ft := field.Type()
		var (
			kwd_name     = fmt.Sprintf("py_kwd_%03d", i+1)
			cgo_fsetname = fmt.Sprintf("_cffi_helper.lib.cgo_func_%[1]s_setter_%[2]d", s.sym.id, i+1)
			ifield       = newVar(pkg, ft, field.Name(), "ret", "")
		)

		g.wrapper.Printf("%[1]s = None\n", kwd_name)
		g.wrapper.Printf("if  %[1]d < nargs:\n", i)
		g.wrapper.Indent()
		g.wrapper.Printf("%[1]s = args[%[2]d]\n", kwd_name, i)
		g.wrapper.Outdent()
		g.wrapper.Printf("if %[1]q in kwargs:\n", field.Name())
		g.wrapper.Indent()
		g.wrapper.Printf("%[1]s = kwargs[%[2]q]\n", kwd_name, field.Name())
		g.wrapper.Outdent()
		g.wrapper.Printf("if %[1]s != None:\n", kwd_name)
		g.wrapper.Indent()
		g.wrapper.Printf("if not isinstance(%[1]s, %[2]s):\n", kwd_name, ifield.sym.pysig)
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError(\"invalid type for '%[1]s' attribute\")\n", field.Name())
		g.wrapper.Outdent()
		if ifield.sym.hasConverter() {
			g.wrapper.Printf("c_kwd_%03[1]d = _cffi_helper.cffi_%[2]s(py_kwd_%03[1]d)\n", i+1, ifield.sym.py2c)
			g.wrapper.Printf("%[1]s(self.cgopy, c_kwd_%03d)\n", cgo_fsetname, i+1)
		} else {
			g.wrapper.Printf("%[1]s(self.cgopy, %[2]s)\n", cgo_fsetname, kwd_name)
		}
		g.wrapper.Outdent()
	}
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

func (g *cffiGen) genStructMembers(s Struct) {
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

func (g *cffiGen) genStructMemberGetter(s Struct, i int, f types.Object) {
	pkg := s.Package()
	ft := f.Type()
	var (
		cgo_fgetname = fmt.Sprintf("_cffi_helper.lib.cgo_func_%[1]s_getter_%[2]d", s.sym.id, i+1)
		ifield       = newVar(pkg, ft, f.Name(), "ret", "")
		results      = []*Var{ifield}
	)

	g.wrapper.Printf("@property\n")
	g.wrapper.Printf("def %[1]s(self):\n", f.Name())
	g.wrapper.Indent()
	g.wrapper.Printf("cret =  %[1]s(self.cgopy)\n", cgo_fgetname)
	switch len(results) {
	case 1:
		ret := results[0]
		if ret.sym.hasConverter() {
			g.wrapper.Printf("ret = _cffi_helper.cffi_%[1]s(cret)\n", ret.sym.c2py)
			g.wrapper.Printf("return ret\n")
		} else {
			g.wrapper.Printf("return cret\n")
		}
	default:
		panic("bind: impossible")
	}
	g.wrapper.Printf("\n")
	g.wrapper.Outdent()
}

func (g *cffiGen) genStructMemberSetter(s Struct, i int, f types.Object) {
	pkg := s.Package()
	ft := f.Type()
	var (
		cgo_fsetname = fmt.Sprintf("_cffi_helper.lib.cgo_func_%[1]s_setter_%[2]d", s.sym.id, i+1)
		ifield       = newVar(pkg, ft, f.Name(), "ret", "")
	)
	g.wrapper.Printf("@%[1]s.setter\n", f.Name())
	g.wrapper.Printf("def %[1]s(self, value):\n", f.Name())
	g.wrapper.Indent()
	if ifield.sym.hasConverter() {
		g.wrapper.Printf("c_value = _cffi_helper.cffi_%[1]s(value)\n", ifield.sym.py2c)
		g.wrapper.Printf("%[1]s(self.cgopy, c_value)\n", cgo_fsetname)
	} else {
		g.wrapper.Printf("%[1]s(self.cgopy, value)\n", cgo_fsetname)
	}
	g.wrapper.Printf("\n")
	g.wrapper.Outdent()
}

func (g *cffiGen) genStructMethods(s Struct) {
	//pkgname := s.Package().Name()
	for _, m := range s.meths {
		g.genMethod(s, m)
	}
}

func (g *cffiGen) genStructTPRepr(s Struct) {
	g.wrapper.Printf("def __repr__(self):\n")
	g.wrapper.Indent()
	g.wrapper.Printf("cret = _cffi_helper.lib.cgo_func_%[1]s_str(self.cgopy)\n", s.sym.id)
	g.wrapper.Printf("ret = _cffi_helper.cffi_cgopy_cnv_c2py_string(cret)\n")
	g.wrapper.Printf("return ret\n")
	g.wrapper.Printf("\n")
	g.wrapper.Outdent()
}

func (g *cffiGen) genStructTPStr(s Struct) {
	g.wrapper.Printf("def __str__(self):\n")
	g.wrapper.Indent()
	g.wrapper.Printf("cret = _cffi_helper.lib.cgo_func_%[1]s_str(self.cgopy)\n", s.sym.id)
	g.wrapper.Printf("ret = _cffi_helper.cffi_cgopy_cnv_c2py_string(cret)\n")
	g.wrapper.Printf("return ret\n")
	g.wrapper.Printf("\n")
	g.wrapper.Outdent()
}
