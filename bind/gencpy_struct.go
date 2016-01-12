// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

func (g *cpyGen) genStructInit(cpy Type) {
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

func (g *cpyGen) genStructMembers(cpy Type) {
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

func (g *cpyGen) genStructMemberGetter(cpy Type, i int, f types.Object) {
	pkg := cpy.Package()
	ft := f.Type()
	var (
		cpy_fgetname = fmt.Sprintf("cpy_func_%[1]s_getter_%[2]d", cpy.sym.id, i+1)
		ifield       = newVar(pkg, ft, f.Name(), "ret", "")
		results      = []*Var{ifield}
	)

	recv := newVar(cpy.pkg, cpy.GoType(), "self", cpy.GoName(), "")

	fget := Func{
		pkg:  cpy.pkg,
		sig:  newSignature(cpy.pkg, recv, nil, results),
		typ:  nil,
		name: f.Name(),
		desc: pkg.ImportPath() + "." + cpy.GoName() + "." + f.Name() + ".get",
		id:   cpy.ID() + "_" + f.Name() + "_get",
		doc:  "",
		ret:  ft,
		err:  false,
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
	g.genFuncBody(fget)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genStructMemberSetter(cpy Type, i int, f types.Object) {
	var (
		pkg          = cpy.Package()
		ft           = f.Type()
		ifield       = newVar(pkg, ft, f.Name(), "ret", "")
		cpy_fsetname = fmt.Sprintf("cpy_func_%[1]s_setter_%[2]d", cpy.sym.id, i+1)
		params       = []*Var{ifield}
		recv         = newVar(cpy.pkg, cpy.GoType(), "self", cpy.GoName(), "")
	)

	fset := Func{
		pkg:  cpy.pkg,
		sig:  newSignature(cpy.pkg, recv, params, nil),
		typ:  nil,
		name: f.Name(),
		desc: pkg.ImportPath() + "." + cpy.GoName() + "." + f.Name() + ".set",
		id:   cpy.ID() + "_" + f.Name() + "_set",
		doc:  "",
		ret:  nil,
		err:  false,
	}

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

	// create in/out seq-buffers
	g.impl.Printf("cgopy_seq_buffer ibuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("cgopy_seq_buffer obuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("\n")

	// fill input seq-buffer
	g.genWrite("self->cgopy", "ibuf", cpy.sym.GoType())
	g.genWrite("c_"+ifield.Name(), "ibuf", ifield.GoType())
	g.impl.Printf("\n")

	g.impl.Printf("cgopy_seq_send(%q, %d, ibuf->buf, ibuf->len, &obuf->buf, &obuf->len);\n\n",
		fset.Descriptor(),
		uhash(fset.id),
	)

	g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
	g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")

	g.impl.Printf("return 0;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}
