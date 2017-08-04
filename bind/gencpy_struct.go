// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

func (g *cpyGen) genStruct(cpy Struct) {
	pkgname := cpy.Package().Name()

	//fmt.Printf("obj: %#v\ntyp: %#v\n", obj, typ)
	g.decl.Printf("\n/* --- decls for struct %s.%v --- */\n", pkgname, cpy.GoName())
	g.decl.Printf("typedef void* %s;\n\n", cpy.sym.cgoname)
	g.decl.Printf("/* Python type for struct %s.%v\n", pkgname, cpy.GoName())
	g.decl.Printf(" */\ntypedef struct {\n")
	g.decl.Indent()
	g.decl.Printf("PyObject_HEAD\n")
	g.decl.Printf("%[1]s cgopy; /* unsafe.Pointer to %[2]s */\n",
		cpy.sym.cgoname,
		cpy.ID(),
	)
	g.decl.Printf("gopy_efacefunc eface;\n")
	g.decl.Outdent()
	g.decl.Printf("} %s;\n", cpy.sym.cpyname)
	g.decl.Printf("\n\n")

	g.impl.Printf("\n\n/* --- impl for %s.%v */\n\n", pkgname, cpy.GoName())

	g.genStructNew(cpy)
	g.genStructDealloc(cpy)
	g.genStructInit(cpy)
	g.genStructMembers(cpy)
	g.genStructMethods(cpy)

	g.genStructProtocols(cpy)

	g.impl.Printf("static PyTypeObject %sType = {\n", cpy.sym.cpyname)
	g.impl.Indent()
	g.impl.Printf("PyObject_HEAD_INIT(NULL)\n")
	g.impl.Printf("0,\t/*ob_size*/\n")
	g.impl.Printf("\"%s\",\t/*tp_name*/\n", cpy.GoName())
	g.impl.Printf("sizeof(%s),\t/*tp_basicsize*/\n", cpy.sym.cpyname)
	g.impl.Printf("0,\t/*tp_itemsize*/\n")
	g.impl.Printf("(destructor)%s_dealloc,\t/*tp_dealloc*/\n", cpy.sym.cpyname)
	g.impl.Printf("0,\t/*tp_print*/\n")
	g.impl.Printf("0,\t/*tp_getattr*/\n")
	g.impl.Printf("0,\t/*tp_setattr*/\n")
	g.impl.Printf("0,\t/*tp_compare*/\n")
	g.impl.Printf("0,\t/*tp_repr*/\n")
	g.impl.Printf("0,\t/*tp_as_number*/\n")
	g.impl.Printf("0,\t/*tp_as_sequence*/\n")
	g.impl.Printf("0,\t/*tp_as_mapping*/\n")
	g.impl.Printf("0,\t/*tp_hash */\n")
	g.impl.Printf("0,\t/*tp_call*/\n")
	g.impl.Printf("cpy_func_%s_tp_str,\t/*tp_str*/\n", cpy.sym.id)
	g.impl.Printf("0,\t/*tp_getattro*/\n")
	g.impl.Printf("0,\t/*tp_setattro*/\n")
	g.impl.Printf("0,\t/*tp_as_buffer*/\n")
	g.impl.Printf("Py_TPFLAGS_DEFAULT,\t/*tp_flags*/\n")
	g.impl.Printf("%q,\t/* tp_doc */\n", cpy.Doc())
	g.impl.Printf("0,\t/* tp_traverse */\n")
	g.impl.Printf("0,\t/* tp_clear */\n")
	g.impl.Printf("0,\t/* tp_richcompare */\n")
	g.impl.Printf("0,\t/* tp_weaklistoffset */\n")
	g.impl.Printf("0,\t/* tp_iter */\n")
	g.impl.Printf("0,\t/* tp_iternext */\n")
	g.impl.Printf("%s_methods,             /* tp_methods */\n", cpy.sym.cpyname)
	g.impl.Printf("0,\t/* tp_members */\n")
	g.impl.Printf("%s_getsets,\t/* tp_getset */\n", cpy.sym.cpyname)
	g.impl.Printf("0,\t/* tp_base */\n")
	g.impl.Printf("0,\t/* tp_dict */\n")
	g.impl.Printf("0,\t/* tp_descr_get */\n")
	g.impl.Printf("0,\t/* tp_descr_set */\n")
	g.impl.Printf("0,\t/* tp_dictoffset */\n")
	g.impl.Printf("(initproc)cpy_func_%s_init,      /* tp_init */\n", cpy.sym.id)
	g.impl.Printf("0,                         /* tp_alloc */\n")
	g.impl.Printf("cpy_func_%s_new,\t/* tp_new */\n", cpy.sym.id)
	g.impl.Outdent()
	g.impl.Printf("};\n\n")

	g.genStructConverters(cpy)
	g.genStructTypeCheck(cpy)

}

func (g *cpyGen) genStructNew(cpy Struct) {
	g.genTypeNew(cpy.sym)
}

func (g *cpyGen) genStructDealloc(cpy Struct) {
	g.genTypeDealloc(cpy.sym)
}

func (g *cpyGen) genStructInit(cpy Struct) {
	pkgname := cpy.Package().Name()

	g.decl.Printf("\n/* tp_init for %s.%v */\n", pkgname, cpy.GoName())
	g.decl.Printf(
		"static int\ncpy_func_%[1]s_init(%[2]s *self, PyObject *args, PyObject *kwds);\n",
		cpy.sym.id,
		cpy.sym.cpyname,
	)

	g.impl.Printf("\n/* tp_init */\n")
	g.impl.Printf(
		"static int\ncpy_func_%[1]s_init(%[2]s *self, PyObject *args, PyObject *kwds) {\n",
		cpy.sym.id,
		cpy.sym.cpyname,
	)
	g.impl.Indent()

	numFields := cpy.Struct().NumFields()
	numPublic := numFields
	for i := 0; i < cpy.Struct().NumFields(); i++ {
		f := cpy.Struct().Field(i)
		if !f.Exported() {
			numPublic--
			continue
		}
	}

	if numPublic > 0 {
		kwds := make(map[string]int)
		g.impl.Printf("static char *kwlist[] = {\n")
		g.impl.Indent()
		for i := 0; i < numFields; i++ {
			field := cpy.Struct().Field(i)
			if !field.Exported() {
				continue
			}
			kwds[field.Name()] = i
			g.impl.Printf("%q, /* py_kwd_%03d */\n", field.Name(), i)
		}
		g.impl.Printf("NULL\n")
		g.impl.Outdent()
		g.impl.Printf("};\n")

		for i := 0; i < numFields; i++ {
			field := cpy.Struct().Field(i)
			if !field.Exported() {
				continue
			}
			g.impl.Printf("PyObject *py_kwd_%03d = NULL;\n", i)
		}

		g.impl.Printf("Py_ssize_t nkwds = (kwds != NULL) ? PyDict_Size(kwds) : 0;\n")
		g.impl.Printf("Py_ssize_t nargs = (args != NULL) ? PySequence_Size(args) : 0;\n")
		g.impl.Printf("if ((nkwds + nargs) > %d) {\n", numPublic)
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
		g.impl.Printf("\"%s.__init__ takes at most %d argument(s)\");\n",
			cpy.GoName(),
			numPublic,
		)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", cpy.sym.cpyname)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("if (!PyArg_ParseTupleAndKeywords(args, kwds, ")
		format := []string{"|"}
		addrs := []string{}
		for i := 0; i < numFields; i++ {
			field := cpy.Struct().Field(i)
			if !field.Exported() {
				continue
			}
			format = append(format, "O")
			addrs = append(addrs, fmt.Sprintf("&py_kwd_%03d", i))
		}
		g.impl.Printf("%q, kwlist, %s)) {\n",
			strings.Join(format, ""),
			strings.Join(addrs, ", "),
		)
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", cpy.sym.cpyname)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		for i := 0; i < numFields; i++ {
			field := cpy.Struct().Field(i)
			if !field.Exported() {
				continue
			}
			g.impl.Printf("if (py_kwd_%03d != NULL) {\n", i)
			g.impl.Indent()
			g.impl.Printf(
				"if (cpy_func_%[1]s_setter_%[2]d(self, py_kwd_%03[3]d, NULL)) {\n",
				cpy.sym.id,
				i+1,
				i,
			)
			g.impl.Indent()
			g.impl.Printf("goto cpy_label_%s_init_fail;\n", cpy.sym.cpyname)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.impl.Outdent()
			g.impl.Printf("}\n\n")
		}
	}
	g.impl.Printf("return 0;\n")
	g.impl.Outdent()
	g.impl.Printf("\ncpy_label_%s_init_fail:\n", cpy.sym.cpyname)
	g.impl.Indent()
	for i := 0; i < numFields; i++ {
		field := cpy.Struct().Field(i)
		if !field.Exported() {
			continue
		}
		g.impl.Printf("Py_XDECREF(py_kwd_%03d);\n", i)
	}
	g.impl.Printf("\nreturn -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genStructMembers(cpy Struct) {
	pkgname := cpy.Package().Name()
	typ := cpy.Struct()

	g.decl.Printf("\n/* tp_getset for %s.%v */\n", pkgname, cpy.GoName())
	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() {
			continue
		}
		g.genStructMemberGetter(cpy, i, f)
		g.genStructMemberSetter(cpy, i, f)
	}

	g.impl.Printf("\n/* tp_getset for %s.%v */\n", pkgname, cpy.GoName())
	g.impl.Printf("static PyGetSetDef %s_getsets[] = {\n", cpy.sym.cpyname)
	g.impl.Indent()
	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() {
			continue
		}
		doc := "doc for " + f.Name() // FIXME(sbinet) retrieve doc for fields
		g.impl.Printf("{%q, ", f.Name())
		g.impl.Printf("(getter)cpy_func_%[1]s_getter_%[2]d, ", cpy.sym.id, i+1)
		g.impl.Printf("(setter)cpy_func_%[1]s_setter_%[2]d, ", cpy.sym.id, i+1)
		g.impl.Printf("%q, NULL},\n", doc)
	}
	g.impl.Printf("{NULL} /* Sentinel */\n")
	g.impl.Outdent()
	g.impl.Printf("};\n\n")
}

func (g *cpyGen) genStructMemberGetter(cpy Struct, i int, f types.Object) {
	pkg := cpy.Package()
	ft := f.Type()
	var (
		cgo_fgetname = fmt.Sprintf("cgo_func_%[1]s_getter_%[2]d", cpy.sym.id, i+1)
		cpy_fgetname = fmt.Sprintf("cpy_func_%[1]s_getter_%[2]d", cpy.sym.id, i+1)
		ifield       = newVar(pkg, ft, f.Name(), "ret", "")
		results      = []*Var{ifield}
	)

	if needWrapType(ft) {
		g.decl.Printf("\n/* wrapper for field %s.%s.%s */\n",
			pkg.Name(),
			cpy.GoName(),
			f.Name(),
		)
		g.decl.Printf("typedef void* %[1]s_field_%d;\n", cpy.sym.cgoname, i+1)
	}

	g.decl.Printf("\n/* getter for %[1]s.%[2]s.%[3]s */\n",
		pkg.Name(), cpy.sym.goname, f.Name(),
	)
	g.decl.Printf("static PyObject*\n")
	g.decl.Printf(
		"%[2]s(%[1]s *self, void *closure); /* %[3]s */\n",
		cpy.sym.cpyname,
		cpy_fgetname,
		f.Name(),
	)

	g.impl.Printf("\n/* getter for %[1]s.%[2]s.%[3]s */\n",
		pkg.Name(), cpy.sym.goname, f.Name(),
	)
	g.impl.Printf("static PyObject*\n")
	g.impl.Printf(
		"%[2]s(%[1]s *self, void *closure) /* %[3]s */ {\n",
		cpy.sym.cpyname,
		cpy_fgetname,
		f.Name(),
	)
	g.impl.Indent()

	g.impl.Printf("PyObject *o = NULL;\n")
	ftname := g.pkg.syms.symtype(ft).cgoname
	if needWrapType(ft) {
		ftname = fmt.Sprintf("%[1]s_field_%d", cpy.sym.cgoname, i+1)
	}
	g.impl.Printf(
		"%[1]s c_ret = %[2]s(self->cgopy); /*wrap*/\n",
		ftname,
		cgo_fgetname,
	)

	{
		format := []string{}
		funcArgs := []string{}
		switch len(results) {
		case 1:
			ret := results[0]
			ret.name = "ret"
			pyfmt, pyaddrs := ret.getArgBuildValue()
			format = append(format, pyfmt)
			funcArgs = append(funcArgs, pyaddrs...)
		default:
			panic("bind: impossible")
		}
		g.impl.Printf("o = Py_BuildValue(%q, %s);\n",
			strings.Join(format, ""),
			strings.Join(funcArgs, ", "),
		)
	}

	g.impl.Printf("return o;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

}

func (g *cpyGen) genStructMemberSetter(cpy Struct, i int, f types.Object) {
	var (
		pkg          = cpy.Package()
		ft           = f.Type()
		self         = newVar(pkg, cpy.GoType(), cpy.GoName(), "self", "")
		ifield       = newVar(pkg, ft, f.Name(), "ret", "")
		cgo_fsetname = fmt.Sprintf("cgo_func_%[1]s_setter_%[2]d", cpy.sym.id, i+1)
		cpy_fsetname = fmt.Sprintf("cpy_func_%[1]s_setter_%[2]d", cpy.sym.id, i+1)
	)

	g.decl.Printf("\n/* setter for %[1]s.%[2]s.%[3]s */\n",
		pkg.Name(), cpy.sym.goname, f.Name(),
	)

	g.decl.Printf("static int\n")
	g.decl.Printf(
		"%[2]s(%[1]s *self, PyObject *value, void *closure);\n",
		cpy.sym.cpyname,
		cpy_fsetname,
	)

	g.impl.Printf("\n/* setter for %[1]s.%[2]s.%[3]s */\n",
		pkg.Name(), cpy.sym.goname, f.Name(),
	)
	g.impl.Printf("static int\n")
	g.impl.Printf(
		"%[2]s(%[1]s *self, PyObject *value, void *closure) {\n",
		cpy.sym.cpyname,
		cpy_fsetname,
	)
	g.impl.Indent()

	ifield.genDecl(g.impl)
	g.impl.Printf("if (value == NULL) {\n")
	g.impl.Indent()
	g.impl.Printf(
		"PyErr_SetString(PyExc_TypeError, \"cannot delete '%[1]s' attribute\");\n",
		f.Name(),
	)
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	g.impl.Printf("if (!%s) {\n", fmt.Sprintf(ifield.sym.pychk, "value"))
	g.impl.Indent()
	g.impl.Printf(
		"PyErr_SetString(PyExc_TypeError, \"invalid type for '%[1]s' attribute\");\n",
		f.Name(),
	)
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	g.impl.Printf("if (!%[1]s(value, &c_ret)) {\n", ifield.sym.py2c)
	g.impl.Indent()
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	g.impl.Printf("%[1]s((%[2]s)(self->cgopy), c_%[3]s);\n",
		cgo_fsetname,
		self.CGoType(),
		ifield.Name(),
	)

	g.impl.Printf("return 0;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genStructMethods(cpy Struct) {

	pkgname := cpy.Package().Name()

	g.decl.Printf("\n/* methods for %s.%s */\n", pkgname, cpy.GoName())
	typ := cpy.sym.GoType().(*types.Named)
	for i := 0; i < typ.NumMethods(); i++ {
		m := typ.Method(i)
		if !m.Exported() {
			continue
		}
		mname := types.ObjectString(m, nil)
		msym := g.pkg.syms.sym(mname)
		if msym == nil {
			panic(fmt.Errorf(
				"gopy: could not find symbol for %q",
				m.FullName(),
			))
		}
		g._genFunc(cpy.sym, msym)
	}

	g.impl.Printf("\n/* methods for %s.%s */\n", pkgname, cpy.GoName())
	g.impl.Printf("static PyMethodDef %s_methods[] = {\n", cpy.sym.cpyname)
	g.impl.Indent()
	for _, m := range cpy.meths {
		margs := "METH_VARARGS"
		if len(m.Signature().Params()) == 0 {
			margs = "METH_NOARGS"
		}
		g.impl.Printf(
			"{%[1]q, (PyCFunction)cpy_func_%[2]s, %[3]s, %[4]q},\n",
			m.GoName(),
			m.ID(),
			margs,
			m.Doc(),
		)
	}
	g.impl.Printf("{NULL} /* sentinel */\n")
	g.impl.Outdent()
	g.impl.Printf("};\n\n")
}

func (g *cpyGen) genStructProtocols(cpy Struct) {
	g.genStructTPStr(cpy)
}

func (g *cpyGen) genStructTPStr(cpy Struct) {
	g.genTypeTPStr(cpy.sym)
}

func (g *cpyGen) genStructConverters(cpy Struct) {
	g.genTypeConverter(cpy.sym)
}

func (g *cpyGen) genStructTypeCheck(cpy Struct) {
	g.genTypeTypeCheck(cpy.sym)
}
