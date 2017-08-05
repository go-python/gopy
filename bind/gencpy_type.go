// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

func (g *cpyGen) genType(sym *symbol) {
	if !sym.isType() {
		return
	}
	if sym.isStruct() {
		return
	}
	if sym.isBasic() && !sym.isNamed() {
		return
	}

	g.decl.Printf("\n/* --- decls for type %v --- */\n", sym.gofmt())
	if sym.isBasic() {
		// reach at the underlying type
		btyp := g.pkg.syms.symtype(sym.GoType().Underlying())
		g.decl.Printf("typedef %s %s;\n\n", btyp.cgoname, sym.cgoname)
	} else {
		g.decl.Printf("typedef void* %s;\n\n", sym.cgoname)
	}
	g.decl.Printf("/* Python type for %v\n", sym.gofmt())
	g.decl.Printf(" */\ntypedef struct {\n")
	g.decl.Indent()
	g.decl.Printf("PyObject_HEAD\n")
	if sym.isBasic() {
		g.decl.Printf("%[1]s cgopy; /* value of %[2]s */\n",
			sym.cgoname,
			sym.id,
		)
	} else {
		g.decl.Printf("%[1]s cgopy; /* unsafe.Pointer to %[2]s */\n",
			sym.cgoname,
			sym.id,
		)
	}
	g.decl.Printf("gopy_efacefunc eface;\n")
	g.decl.Outdent()
	g.decl.Printf("} %s;\n", sym.cpyname)
	g.decl.Printf("\n\n")

	g.impl.Printf("\n\n/* --- impl for %s */\n\n", sym.gofmt())

	g.genTypeNew(sym)
	g.genTypeDealloc(sym)
	g.genTypeInit(sym)
	g.genTypeMembers(sym)
	g.genTypeMethods(sym)

	g.genTypeProtocols(sym)

	tpAsBuffer := "0"
	tpAsSequence := "0"
	tpFlags := "Py_TPFLAGS_DEFAULT"
	if sym.isArray() || sym.isSlice() {
		tpAsBuffer = fmt.Sprintf("&%[1]s_tp_as_buffer", sym.cpyname)
		tpAsSequence = fmt.Sprintf("&%[1]s_tp_as_sequence", sym.cpyname)
		switch g.lang {
		case 2:
			tpFlags = fmt.Sprintf(
				"(%s)",
				strings.Join([]string{
					"Py_TPFLAGS_DEFAULT",
					"Py_TPFLAGS_HAVE_NEWBUFFER",
				},
					" |\n ",
				))
		case 3:
		}
	}

	tpCall := "0"
	if sym.isSignature() {
		sig := sym.GoType().Underlying().(*types.Signature)
		if sig.Recv() == nil {
			// only generate tp_call for functions (not methods)
			tpCall = fmt.Sprintf("(ternaryfunc)cpy_func_%[1]s_tp_call", sym.id)
		}
	}

	g.impl.Printf("static PyTypeObject %sType = {\n", sym.cpyname)
	g.impl.Indent()
	g.impl.Printf("PyObject_HEAD_INIT(NULL)\n")
	g.impl.Printf("0,\t/*ob_size*/\n")
	g.impl.Printf("\"%s\",\t/*tp_name*/\n", sym.gofmt())
	g.impl.Printf("sizeof(%s),\t/*tp_basicsize*/\n", sym.cpyname)
	g.impl.Printf("0,\t/*tp_itemsize*/\n")
	g.impl.Printf("(destructor)%s_dealloc,\t/*tp_dealloc*/\n", sym.cpyname)
	g.impl.Printf("0,\t/*tp_print*/\n")
	g.impl.Printf("0,\t/*tp_getattr*/\n")
	g.impl.Printf("0,\t/*tp_setattr*/\n")
	g.impl.Printf("0,\t/*tp_compare*/\n")
	g.impl.Printf("0,\t/*tp_repr*/\n")
	g.impl.Printf("0,\t/*tp_as_number*/\n")
	g.impl.Printf("%s,\t/*tp_as_sequence*/\n", tpAsSequence)
	g.impl.Printf("0,\t/*tp_as_mapping*/\n")
	g.impl.Printf("0,\t/*tp_hash */\n")
	g.impl.Printf("%s,\t/*tp_call*/\n", tpCall)
	g.impl.Printf("cpy_func_%s_tp_str,\t/*tp_str*/\n", sym.id)
	g.impl.Printf("0,\t/*tp_getattro*/\n")
	g.impl.Printf("0,\t/*tp_setattro*/\n")
	g.impl.Printf("%s,\t/*tp_as_buffer*/\n", tpAsBuffer)
	g.impl.Printf("%s,\t/*tp_flags*/\n", tpFlags)
	g.impl.Printf("%q,\t/* tp_doc */\n", sym.doc)
	g.impl.Printf("0,\t/* tp_traverse */\n")
	g.impl.Printf("0,\t/* tp_clear */\n")
	g.impl.Printf("0,\t/* tp_richcompare */\n")
	g.impl.Printf("0,\t/* tp_weaklistoffset */\n")
	g.impl.Printf("0,\t/* tp_iter */\n")
	g.impl.Printf("0,\t/* tp_iternext */\n")
	g.impl.Printf("%s_methods,             /* tp_methods */\n", sym.cpyname)
	g.impl.Printf("0,\t/* tp_members */\n")
	g.impl.Printf("%s_getsets,\t/* tp_getset */\n", sym.cpyname)
	g.impl.Printf("0,\t/* tp_base */\n")
	g.impl.Printf("0,\t/* tp_dict */\n")
	g.impl.Printf("0,\t/* tp_descr_get */\n")
	g.impl.Printf("0,\t/* tp_descr_set */\n")
	g.impl.Printf("0,\t/* tp_dictoffset */\n")
	g.impl.Printf("(initproc)%s_init,      /* tp_init */\n", sym.cpyname)
	g.impl.Printf("0,                         /* tp_alloc */\n")
	g.impl.Printf("cpy_func_%s_new,\t/* tp_new */\n", sym.id)
	g.impl.Outdent()
	g.impl.Printf("};\n\n")

	g.genTypeConverter(sym)
	g.genTypeTypeCheck(sym)
}

func (g *cpyGen) genTypeNew(sym *symbol) {
	g.decl.Printf("\n/* tp_new for %s */\n", sym.gofmt())
	g.decl.Printf(
		"static PyObject*\ncpy_func_%s_new(PyTypeObject *type, PyObject *args, PyObject *kwds);\n",
		sym.id,
	)

	g.impl.Printf("\n/* tp_new */\n")
	g.impl.Printf(
		"static PyObject*\ncpy_func_%s_new(PyTypeObject *type, PyObject *args, PyObject *kwds) {\n",
		sym.id,
	)
	g.impl.Indent()
	g.impl.Printf("%s *self;\n", sym.cpyname)
	g.impl.Printf("self = (%s *)type->tp_alloc(type, 0);\n", sym.cpyname)
	g.impl.Printf("self->cgopy = cgo_func_%s_new();\n", sym.id)
	g.impl.Printf("self->eface = (gopy_efacefunc)cgo_func_%s_eface;\n", sym.id)
	g.impl.Printf("return (PyObject*)self;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeDealloc(sym *symbol) {
	g.decl.Printf("\n/* tp_dealloc for %s */\n", sym.gofmt())
	g.decl.Printf("static void\n%[1]s_dealloc(%[1]s *self);\n",
		sym.cpyname,
	)

	g.impl.Printf("\n/* tp_dealloc for %s */\n", sym.gofmt())
	g.impl.Printf("static void\n%[1]s_dealloc(%[1]s *self) {\n",
		sym.cpyname,
	)
	g.impl.Indent()
	if !sym.isBasic() {
		g.impl.Printf("cgopy_decref((%[1]s)(self->cgopy));\n", sym.cgoname)
	}
	g.impl.Printf("self->ob_type->tp_free((PyObject*)self);\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeInit(sym *symbol) {
	g.decl.Printf("\n/* tp_init for %s */\n", sym.gofmt())
	g.decl.Printf(
		"static int\n%[1]s_init(%[1]s *self, PyObject *args, PyObject *kwds);\n",
		sym.cpyname,
	)

	g.impl.Printf("\n/* tp_init */\n")
	g.impl.Printf(
		"static int\n%[1]s_init(%[1]s *self, PyObject *args, PyObject *kwds) {\n",
		sym.cpyname,
	)
	g.impl.Indent()

	g.impl.Printf("static char *kwlist[] = {\n")
	g.impl.Indent()
	g.impl.Printf("%q,\n", "")
	g.impl.Printf("NULL\n")
	g.impl.Outdent()
	g.impl.Printf("};\n")

	nargs := 1
	g.impl.Printf("PyObject *arg = NULL;\n\n")
	g.impl.Printf("Py_ssize_t nkwds = (kwds != NULL) ? PyDict_Size(kwds) : 0;\n")
	g.impl.Printf("Py_ssize_t nargs = (args != NULL) ? PySequence_Size(args) : 0;\n")
	g.impl.Printf("if ((nkwds + nargs) > %d) {\n", nargs)
	g.impl.Indent()
	g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
	g.impl.Printf("\"%s.__init__ takes at most %d argument(s)\");\n",
		sym.goname,
		nargs,
	)
	g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	g.impl.Printf("if (!PyArg_ParseTupleAndKeywords(args, kwds, ")
	format := []string{"|O"}
	addrs := []string{"&arg"}
	g.impl.Printf("%q, kwlist, %s)) {\n",
		strings.Join(format, ""),
		strings.Join(addrs, ", "),
	)
	g.impl.Indent()
	g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	// FIXME(sbinet) handle slices, arrays and funcs
	switch {
	case sym.isBasic():
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()
		bsym := g.pkg.syms.symtype(sym.GoType().Underlying())
		g.impl.Printf(
			"if (!%s(arg, &self->cgopy)) {\n",
			bsym.py2c,
		)
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Outdent()
		g.impl.Printf("}\n\n")

	case sym.isArray():
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()

		g.impl.Printf("if (!PySequence_Check(arg)) {\n")
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
		g.impl.Printf("\"%s.__init__ takes a sequence as argument\");\n", sym.goname)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		typ := sym.GoType().Underlying().(*types.Array)
		esym := g.pkg.syms.symtype(typ.Elem())
		if esym == nil {
			panic(fmt.Errorf(
				"gopy: could not find symbol for element of %q",
				sym.gofmt(),
			))
		}

		g.impl.Printf("Py_ssize_t len = PySequence_Size(arg);\n")
		g.impl.Printf("if (len == -1) {\n")
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("if (len > %d) {\n", typ.Len())
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_ValueError, ")
		g.impl.Printf("\"%s.__init__ takes a sequence of size at most %d\");\n",
			sym.goname,
			typ.Len(),
		)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("Py_ssize_t i = 0;\n")
		g.impl.Printf("for (i = 0; i < len; i++) {\n")
		g.impl.Indent()
		g.impl.Printf("PyObject *elt = PySequence_GetItem(arg, i);\n")
		g.impl.Printf("if (cpy_func_%[1]s_ass_item(self, i, elt)) {\n", sym.id)
		g.impl.Indent()
		g.impl.Printf("Py_XDECREF(elt);\n")
		g.impl.Printf(
			"PyErr_SetString(PyExc_TypeError, \"invalid type (expected a %s)\");\n",
			esym.goname,
		)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("Py_XDECREF(elt);\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n") // for-loop

		g.impl.Outdent()
		g.impl.Printf("}\n\n") // if-arg

	case sym.isSlice():
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()

		g.impl.Printf("if (!PySequence_Check(arg)) {\n")
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
		g.impl.Printf("\"%s.__init__ takes a sequence as argument\");\n", sym.goname)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("if (!cpy_func_%[1]s_inplace_concat(self, arg)) {\n", sym.id)
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n") // if-arg

	case sym.isMap():
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()
		//g.impl.Printf("//put map __init__ functions here")
		g.impl.Outdent()
		g.impl.Printf("}\n\n") // if-arg

	case sym.isSignature():
		//TODO(sbinet)

	case sym.isInterface():
		//TODO(sbinet): check the argument implements the interface.

	default:
		panic(fmt.Errorf(
			"gopy: tp_init for %s not handled",
			sym.gofmt(),
		))
	}

	g.impl.Printf("return 0;\n")
	g.impl.Outdent()

	g.impl.Printf("\ncpy_label_%s_init_fail:\n", sym.id)
	g.impl.Indent()
	g.impl.Printf("Py_XDECREF(arg);\n")
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeMembers(sym *symbol) {
	g.decl.Printf("\n/* tp_getset for %s */\n", sym.gofmt())
	g.impl.Printf("\n/* tp_getset for %s */\n", sym.gofmt())
	g.impl.Printf("static PyGetSetDef %s_getsets[] = {\n", sym.cpyname)
	g.impl.Indent()
	g.impl.Printf("{NULL} /* Sentinel */\n")
	g.impl.Outdent()
	g.impl.Printf("};\n\n")
}

func (g *cpyGen) genTypeMethods(sym *symbol) {
	g.decl.Printf("\n/* methods for %s */\n", sym.gofmt())
	if sym.isNamed() {
		typ := sym.GoType().(*types.Named)
		for imeth := 0; imeth < typ.NumMethods(); imeth++ {
			m := typ.Method(imeth)
			if !m.Exported() {
				continue
			}
			mname := types.ObjectString(m, nil)
			msym := g.pkg.syms.sym(mname)
			if msym == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for [%[1]T] (%#[1]v) (%[2]s)",
					m.Type(),
					m.Name()+" || "+m.FullName(),
				))
			}
			g._genFunc(sym, msym)
		}
	}
	g.impl.Printf("\n/* methods for %s */\n", sym.gofmt())
	g.impl.Printf("static PyMethodDef %s_methods[] = {\n", sym.cpyname)
	g.impl.Indent()
	if sym.isNamed() {
		typ := sym.GoType().(*types.Named)
		for imeth := 0; imeth < typ.NumMethods(); imeth++ {
			m := typ.Method(imeth)
			if !m.Exported() {
				continue
			}
			mname := types.ObjectString(m, nil)
			msym := g.pkg.syms.sym(mname)
			margs := "METH_VARARGS"
			sig := m.Type().Underlying().(*types.Signature)
			if sig.Params() == nil || sig.Params().Len() <= 0 {
				margs = "METH_NOARGS"
			}
			g.impl.Printf(
				"{%[1]q, (PyCFunction)cpy_func_%[2]s, %[3]s, %[4]q},\n",
				msym.goname,
				msym.id,
				margs,
				msym.doc,
			)
		}
	}
	g.impl.Printf("{NULL} /* sentinel */\n")
	g.impl.Outdent()
	g.impl.Printf("};\n\n")
}

func (g *cpyGen) genTypeProtocols(sym *symbol) {
	g.genTypeTPStr(sym)
	if sym.isSlice() || sym.isArray() {
		g.genTypeTPAsSequence(sym)
		g.genTypeTPAsBuffer(sym)
	}
	if sym.isSignature() {
		g.genTypeTPCall(sym)
	}
}

func (g *cpyGen) genTypeTPStr(sym *symbol) {
	g.decl.Printf("\n/* __str__ support for %[1]s.%[2]s */\n",
		g.pkg.pkg.Name(),
		sym.goname,
	)
	g.decl.Printf(
		"static PyObject*\ncpy_func_%s_tp_str(PyObject *self);\n",
		sym.id,
	)

	g.impl.Printf(
		"static PyObject*\ncpy_func_%s_tp_str(PyObject *self) {\n",
		sym.id,
	)

	g.impl.Indent()
	g.impl.Printf("%[1]s c_self = ((%[2]s*)self)->cgopy;\n",
		sym.cgoname,
		sym.cpyname,
	)
	g.impl.Printf("GoString str = cgo_func_%[1]s_str(c_self);\n",
		sym.id,
	)
	g.impl.Printf("return cgopy_cnv_c2py_string(&str);\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeTPAsSequence(sym *symbol) {
	g.decl.Printf("\n/* sequence support for %s */\n", sym.gofmt())

	var arrlen int64
	var etyp types.Type
	switch typ := sym.GoType().(type) {
	case *types.Array:
		etyp = typ.Elem()
		arrlen = typ.Len()
	case *types.Slice:
		etyp = typ.Elem()
	case *types.Named:
		switch typ := typ.Underlying().(type) {
		case *types.Array:
			etyp = typ.Elem()
			arrlen = typ.Len()
		case *types.Slice:
			etyp = typ.Elem()
		default:
			panic(fmt.Errorf(
				"gopy: unhandled type [%#v] (go=%s kind=%v)",
				typ,
				sym.gofmt(),
				sym.kind,
			))
		}
	default:
		panic(fmt.Errorf(
			"gopy: unhandled type [%#v] (go=%s kind=%v)",
			typ,
			sym.gofmt(),
			sym.kind,
		))
	}
	esym := g.pkg.syms.symtype(etyp)
	if esym == nil {
		panic(fmt.Errorf("gopy: could not retrieve element type of %#v",
			sym,
		))
	}

	switch g.lang {
	case 2:

		g.decl.Printf("\n/* len */\n")
		g.decl.Printf("static Py_ssize_t\ncpy_func_%[1]s_len(%[2]s *self);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* len */\n")
		g.impl.Printf("static Py_ssize_t\ncpy_func_%[1]s_len(%[2]s *self) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		if sym.isArray() {
			g.impl.Printf("return %d;\n", arrlen)
		} else {
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("return slice->len;\n")
		}
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* item */\n")
		g.decl.Printf("static PyObject*\n")
		g.decl.Printf("cpy_func_%[1]s_item(%[2]s *self, Py_ssize_t i);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* item */\n")
		g.impl.Printf("static PyObject*\n")
		g.impl.Printf("cpy_func_%[1]s_item(%[2]s *self, Py_ssize_t i) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		g.impl.Printf("PyObject *pyitem = NULL;\n")
		if sym.isArray() {
			g.impl.Printf("if (i < 0 || i >= %d) {\n", arrlen)
		} else {
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (i < 0 || i >= slice->len) {\n")
		}
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_IndexError, ")
		if sym.isArray() {
			g.impl.Printf("\"array index out of range\");\n")
		} else {
			g.impl.Printf("\"slice index out of range\");\n")
		}
		g.impl.Printf("return NULL;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("%[1]s item = cgo_func_%[2]s_item(self->cgopy, i);\n",
			esym.cgoname,
			sym.id,
		)
		g.impl.Printf("pyitem = %[1]s(&item);\n", esym.c2py)
		g.impl.Printf("return pyitem;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* ass_item */\n")
		g.decl.Printf("static int\n")
		g.decl.Printf("cpy_func_%[1]s_ass_item(%[2]s *self, Py_ssize_t i, PyObject *v);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* ass_item */\n")
		g.impl.Printf("static int\n")
		g.impl.Printf("cpy_func_%[1]s_ass_item(%[2]s *self, Py_ssize_t i, PyObject *v) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		g.impl.Printf("%[1]s c_v;\n", esym.cgoname)
		if sym.isArray() {
			g.impl.Printf("if (i < 0 || i >= %d) {\n", arrlen)
		} else {
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (i < 0 || i >= slice->len) {\n")
		}
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_IndexError, ")
		g.impl.Printf("\"array assignment index out of range\");\n")
		g.impl.Printf("return -1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("if (v == NULL) { return 0; }\n") // FIXME(sbinet): semantics?
		g.impl.Printf("if (!%[1]s(v, &c_v)) { return -1; }\n", esym.py2c)
		g.impl.Printf("cgo_func_%[1]s_ass_item(self->cgopy, i, c_v);\n", sym.id)
		g.impl.Printf("return 0;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		sq_inplace_concat := "0"
		// append
		if sym.isSlice() {
			sq_inplace_concat = fmt.Sprintf(
				"cpy_func_%[1]s_inplace_concat",
				sym.id,
			)

			g.decl.Printf("\n/* append-item */\n")
			g.decl.Printf("static int\n")
			g.decl.Printf("cpy_func_%[1]s_append(%[2]s *self, PyObject *v);\n",
				sym.id,
				sym.cpyname,
			)

			g.impl.Printf("\n/* append-item */\n")
			g.impl.Printf("static int\n")
			g.impl.Printf("cpy_func_%[1]s_append(%[2]s *self, PyObject *v) {\n",
				sym.id,
				sym.cpyname,
			)
			g.impl.Indent()
			g.impl.Printf("%[1]s c_v;\n", esym.cgoname)
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (v == NULL) { return 0; }\n") // FIXME(sbinet): semantics?
			g.impl.Printf("if (!%[1]s(v, &c_v)) { return -1; }\n", esym.py2c)
			g.impl.Printf("cgo_func_%[1]s_append(self->cgopy, c_v);\n", sym.id)
			g.impl.Printf("return 0;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.decl.Printf("\n/* inplace-concat */\n")
			g.decl.Printf("static PyObject*\n")
			g.decl.Printf("cpy_func_%[1]s_inplace_concat(%[2]s *self, PyObject *v);\n",
				sym.id,
				sym.cpyname,
			)

			g.impl.Printf("\n/* inplace-item */\n")
			g.impl.Printf("static PyObject*\n")
			g.impl.Printf("cpy_func_%[1]s_inplace_concat(%[2]s *self, PyObject *v) {\n",
				sym.id,
				sym.cpyname,
			)
			g.impl.Indent()
			// FIXME(sbinet) do the append in one go?
			g.impl.Printf("if (!PySequence_Check(v)) {\n")
			g.impl.Indent()
			g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
			g.impl.Printf("\"%s.__iadd__ takes a sequence as argument\");\n", sym.goname)
			g.impl.Printf("goto cpy_label_%s_inplace_concat_fail;\n", sym.id)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.impl.Printf("Py_ssize_t len = PySequence_Size(v);\n")
			g.impl.Printf("if (len == -1) {\n")
			g.impl.Indent()
			g.impl.Printf("goto cpy_label_%s_inplace_concat_fail;\n", sym.id)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.impl.Printf("Py_ssize_t i = 0;\n")
			g.impl.Printf("for (i = 0; i < len; i++) {\n")
			g.impl.Indent()
			g.impl.Printf("PyObject *elt = PySequence_GetItem(v, i);\n")
			g.impl.Printf("if (cpy_func_%[1]s_append(self, elt)) {\n", sym.id)
			g.impl.Indent()
			g.impl.Printf("Py_XDECREF(elt);\n")
			g.impl.Printf(
				"PyErr_Format(PyExc_TypeError, \"invalid type (got=%%s, expected a %s)\", Py_TYPE(elt)->tp_name);\n",
				esym.goname,
			)
			g.impl.Printf("goto cpy_label_%s_inplace_concat_fail;\n", sym.id)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			g.impl.Printf("Py_XDECREF(elt);\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n") // for-loop

			g.impl.Printf("return (PyObject*)self;\n")
			g.impl.Outdent()

			g.impl.Printf("\ncpy_label_%s_inplace_concat_fail:\n", sym.id)
			g.impl.Indent()
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

		}

		g.impl.Printf("\n/* tp_as_sequence */\n")
		g.impl.Printf("static PySequenceMethods %[1]s_tp_as_sequence = {\n", sym.cpyname)
		g.impl.Indent()
		g.impl.Printf("(lenfunc)cpy_func_%[1]s_len,\n", sym.id)
		g.impl.Printf("(binaryfunc)0,\n")   // array_concat,               sq_concat
		g.impl.Printf("(ssizeargfunc)0,\n") //array_repeat,                 /*sq_repeat
		g.impl.Printf("(ssizeargfunc)cpy_func_%[1]s_item,\n", sym.id)
		g.impl.Printf("(ssizessizeargfunc)0,\n") // array_slice,             /*sq_slice
		g.impl.Printf("(ssizeobjargproc)cpy_func_%[1]s_ass_item,\n", sym.id)
		g.impl.Printf("(ssizessizeobjargproc)0,\n") //array_ass_slice,      /*sq_ass_slice
		g.impl.Printf("(objobjproc)0,\n")           //array_contains,                 /*sq_contains
		g.impl.Printf("(binaryfunc)%s,\n", sq_inplace_concat)
		g.impl.Printf("(ssizeargfunc)0\n") //array_inplace_repeat          /*sq_inplace_repeat
		g.impl.Outdent()
		g.impl.Printf("};\n\n")

	case 3:
	}
}

func (g *cpyGen) genTypeTPAsBuffer(sym *symbol) {
	g.decl.Printf("\n/* buffer support for %s */\n", sym.gofmt())

	g.decl.Printf("\n/* __get_buffer__ impl for %s */\n", sym.gofmt())
	g.decl.Printf("static int\n")
	g.decl.Printf(
		"cpy_func_%[1]s_getbuffer(PyObject *self, Py_buffer *view, int flags);\n",
		sym.id,
	)

	var esize int64
	var arrlen int64
	esym := g.pkg.syms.symtype(sym.GoType())
	switch o := sym.GoType().(type) {
	case *types.Array:
		esize = g.pkg.sz.Sizeof(o.Elem())
		arrlen = o.Len()
	case *types.Slice:
		esize = g.pkg.sz.Sizeof(o.Elem())
	}

	g.impl.Printf("\n/* __get_buffer__ impl for %s */\n", sym.gofmt())
	g.impl.Printf("static int\n")
	g.impl.Printf(
		"cpy_func_%[1]s_getbuffer(PyObject *self, Py_buffer *view, int flags) {\n",
		sym.id,
	)
	g.impl.Indent()
	g.impl.Printf("if (view == NULL) {\n")
	g.impl.Indent()
	g.impl.Printf("PyErr_SetString(PyExc_ValueError, ")
	g.impl.Printf("\"NULL view in getbuffer\");\n")
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
	g.impl.Printf("%[1]s *py = (%[1]s*)self;\n", sym.cpyname)
	if sym.isArray() {
		g.impl.Printf("void *array = (void*)(py->cgopy);\n")
		g.impl.Printf("view->obj = (PyObject*)py;\n")
		g.impl.Printf("view->buf = (void*)array;\n")
		g.impl.Printf("view->len = %d;\n", arrlen)
		g.impl.Printf("view->readonly = 0;\n")
		g.impl.Printf("view->itemsize = %d;\n", esize)
		g.impl.Printf("view->format = %q;\n", esym.pybuf)
		g.impl.Printf("view->ndim = 1;\n")
		g.impl.Printf("view->shape = (Py_ssize_t*)&view->len;\n")
	} else {
		g.impl.Printf("GoSlice *slice = (GoSlice*)(py->cgopy);\n")
		g.impl.Printf("view->obj = (PyObject*)py;\n")
		g.impl.Printf("view->buf = (void*)slice->data;\n")
		g.impl.Printf("view->len = slice->len;\n")
		g.impl.Printf("view->readonly = 0;\n")
		g.impl.Printf("view->itemsize = %d;\n", esize)
		g.impl.Printf("view->format = %q;\n", esym.pybuf)
		g.impl.Printf("view->ndim = 1;\n")
		g.impl.Printf("view->shape = (Py_ssize_t*)&slice->len;\n")
	}
	g.impl.Printf("view->strides = &view->itemsize;\n")
	g.impl.Printf("view->suboffsets = NULL;\n")
	g.impl.Printf("view->internal = NULL;\n")

	g.impl.Printf("\nPy_INCREF(py);\n")
	g.impl.Printf("return 0;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	switch g.lang {
	case 2:
		g.decl.Printf("\n/* readbuffer */\n")
		g.decl.Printf("static Py_ssize_t\n")
		g.decl.Printf(
			"cpy_func_%[1]s_readbuffer(%[2]s *self, Py_ssize_t index, const void **ptr);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* readbuffer */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf(
			"cpy_func_%[1]s_readbuffer(%[2]s *self, Py_ssize_t index, const void **ptr) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		g.impl.Printf("if (index != 0) {\n")
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_SystemError, ")
		g.impl.Printf("\"Accessing non-existent array segment\");\n")
		g.impl.Printf("return -1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		if sym.isArray() {
			g.impl.Printf("*ptr = (void*)self->cgopy;\n")
			g.impl.Printf("return %d;\n", arrlen)
		} else {
			g.impl.Printf("GoSlice *slice = (GoSlice*)self->cgopy;\n")
			g.impl.Printf("*ptr = (void*)slice->data;\n")
			g.impl.Printf("return slice->len;\n")
		}
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* writebuffer */\n")
		g.decl.Printf("static Py_ssize_t\n")
		g.decl.Printf(
			"cpy_func_%[1]s_writebuffer(%[2]s *self, Py_ssize_t segment, void **ptr);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* writebuffer */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf(
			"cpy_func_%[1]s_writebuffer(%[2]s *self, Py_ssize_t segment, void **ptr) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		g.impl.Printf("return cpy_func_%[1]s_readbuffer(self, segment, (const void**)ptr);\n",
			sym.id,
		)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* segcount */\n")
		g.decl.Printf("static Py_ssize_t\n")
		g.decl.Printf("cpy_func_%[1]s_segcount(%[2]s *self, Py_ssize_t *lenp);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* segcount */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf("cpy_func_%[1]s_segcount(%[2]s *self, Py_ssize_t *lenp) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		if sym.isArray() {
			g.impl.Printf("if (lenp) { *lenp = %d; }\n", arrlen)
		} else {
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (lenp) { *lenp = slice->len; }\n")
		}
		g.impl.Printf("return 1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* charbuffer */\n")
		g.decl.Printf("static Py_ssize_t\n")
		g.decl.Printf("cpy_func_%[1]s_charbuffer(%[2]s *self, Py_ssize_t segment, const char **ptr);\n",
			sym.id,
			sym.cpyname,
		)

		g.impl.Printf("\n/* charbuffer */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf("cpy_func_%[1]s_charbuffer(%[2]s *self, Py_ssize_t segment, const char **ptr) {\n",
			sym.id,
			sym.cpyname,
		)
		g.impl.Indent()
		g.impl.Printf("return cpy_func_%[1]s_readbuffer(self, segment, (const void**)ptr);\n", sym.id)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

	case 3:
		// no-op
	}

	switch g.lang {
	case 2:
		g.impl.Printf("\n/* tp_as_buffer */\n")
		g.impl.Printf("static PyBufferProcs %[1]s_tp_as_buffer = {\n", sym.cpyname)
		g.impl.Indent()
		g.impl.Printf("(readbufferproc)cpy_func_%[1]s_readbuffer,\n", sym.id)
		g.impl.Printf("(writebufferproc)cpy_func_%[1]s_writebuffer,\n", sym.id)
		g.impl.Printf("(segcountproc)cpy_func_%[1]s_segcount,\n", sym.id)
		g.impl.Printf("(charbufferproc)cpy_func_%[1]s_charbuffer,\n", sym.id)
		g.impl.Printf("(getbufferproc)cpy_func_%[1]s_getbuffer,\n", sym.id)
		g.impl.Printf("(releasebufferproc)0,\n")
		g.impl.Outdent()
		g.impl.Printf("};\n\n")
	case 3:

		g.impl.Printf("\n/* tp_as_buffer */\n")
		g.impl.Printf("static PyBufferProcs %[1]s_tp_as_buffer = {\n", sym.cpyname)
		g.impl.Indent()
		g.impl.Printf("(getbufferproc)cpy_func_%[1]s_getbuffer,\n", sym.id)
		g.impl.Printf("(releasebufferproc)0,\n")
		g.impl.Outdent()
		g.impl.Printf("};\n\n")
	}
}

func (g *cpyGen) genTypeTPCall(sym *symbol) {

	if !sym.isSignature() {
		return
	}

	sig := sym.GoType().Underlying().(*types.Signature)
	if sig.Recv() != nil {
		// don't generate tp_call for methods.
		return
	}

	g.decl.Printf("\n/* tp_call */\n")
	g.decl.Printf("static PyObject *\n")
	g.decl.Printf(
		"cpy_func_%[1]s_tp_call(%[2]s *self, PyObject *args, PyObject *other);\n",
		sym.id,
		sym.cpyname,
	)

	g.impl.Printf("\n/* tp_call */\n")
	g.impl.Printf("static PyObject *\n")
	g.impl.Printf(
		"cpy_func_%[1]s_tp_call(%[2]s *self, PyObject *args, PyObject *other) {\n",
		sym.id,
		sym.cpyname,
	)
	g.impl.Indent()

	//TODO(sbinet): refactor/consolidate with genFuncBody

	funcArgs := []string{"self->cgopy"}
	args := newVarsFrom(g.pkg, sig.Params())
	res := newVarsFrom(g.pkg, sig.Results())

	for _, arg := range args {
		arg.genDecl(g.impl)
		funcArgs = append(funcArgs, arg.getFuncArg())
	}

	if len(res) > 0 {
		switch len(res) {
		case 1:
			ret := res[0]
			ret.genRetDecl(g.impl)
		default:
			g.impl.Printf(
				"struct cgo_func_%[1]s_tp_call_return c_gopy_ret;\n",
				sym.id,
			)
		}
	}

	g.impl.Printf("\n")

	if len(args) > 0 {
		g.impl.Printf("if (!PyArg_ParseTuple(args, ")
		format := []string{}
		pyaddrs := []string{}
		for _, arg := range args {
			pyfmt, addr := arg.getArgParse()
			format = append(format, pyfmt)
			pyaddrs = append(pyaddrs, addr...)
		}
		g.impl.Printf("%q, %s)) {\n", strings.Join(format, ""), strings.Join(pyaddrs, ", "))
		g.impl.Indent()
		g.impl.Printf("return NULL;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
	}

	if len(args) > 0 {
		for _, arg := range args {
			arg.genFuncPreamble(g.impl)
		}
		g.impl.Printf("\n")
	}

	if len(res) > 0 {
		g.impl.Printf("c_gopy_ret = ")
	}
	g.impl.Printf("cgo_func_%[1]s_call(%[2]s);\n", sym.id, strings.Join(funcArgs, ", "))

	g.impl.Printf("\n")

	if len(res) <= 0 {
		g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		return
	}

	if hasError(sig) {
		switch len(res) {
		case 1:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(c_gopy_ret)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(c_gopy_ret);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			return

		case 2:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(c_gopy_ret.r1)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(c_gopy_ret.r1);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			ret := res[0]
			sret := g.pkg.syms.symtype(ret.GoType())
			g.impl.Printf("return %s(&c_gopy_ret.r0);\n", sret.c2py)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			return

		default:
			panic(fmt.Errorf(
				"bind: function/method with more than 2 results not supported! (%s)",
				sym.id,
			))
		}
	}

	ret := res[0]
	sret := g.pkg.syms.symtype(ret.GoType())
	g.impl.Printf("return %s(&c_gopy_ret);\n", sret.c2py)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeConverter(sym *symbol) {
	g.decl.Printf("\n/* converters for %s - %s */\n",
		sym.id,
		sym.goname,
	)
	g.decl.Printf("static int\n")
	g.decl.Printf("cgopy_cnv_py2c_%[1]s(PyObject *o, %[2]s *addr);\n",
		sym.id,
		sym.cgoname,
	)
	g.decl.Printf("static PyObject*\n")
	g.decl.Printf("cgopy_cnv_c2py_%[1]s(%[2]s *addr);\n\n",
		sym.id,
		sym.cgoname,
	)

	g.impl.Printf("static int\n")
	g.impl.Printf("cgopy_cnv_py2c_%[1]s(PyObject *o, %[2]s *addr) {\n",
		sym.id,
		sym.cgoname,
	)
	g.impl.Indent()
	g.impl.Printf("%s *self = NULL;\n", sym.cpyname)
	if sym.isInterface() {
		g.impl.Printf("if (%s) {\n", fmt.Sprintf(sym.pychk, "o"))
		g.impl.Indent()
		g.impl.Printf("self = (%s *)o;\n", sym.cpyname)
		g.impl.Printf("*addr = self->cgopy;\n")
		g.impl.Printf("return 1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("GoInterface *iface = (GoInterface*)(addr);\n")
		g.impl.Printf("*iface = ((gopy_object*)o)->eface((gopy_object*)o);\n")
		g.impl.Printf("return 1;\n")
	} else {
		g.impl.Printf("self = (%s *)o;\n", sym.cpyname)
		g.impl.Printf("*addr = self->cgopy;\n")
		g.impl.Printf("return 1;\n")
	}
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	g.impl.Printf("static PyObject*\n")
	g.impl.Printf("cgopy_cnv_c2py_%[1]s(%[2]s *addr) {\n", sym.id, sym.cgoname)
	g.impl.Indent()
	g.impl.Printf("PyObject *o = cpy_func_%[1]s_new(&%[2]sType, 0, 0);\n",
		sym.id,
		sym.cpyname,
	)
	g.impl.Printf("if (o == NULL) {\n")
	g.impl.Indent()
	g.impl.Printf("return NULL;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n")
	g.impl.Printf("((%[1]s*)o)->cgopy = *addr;\n", sym.cpyname)
	g.impl.Printf("return o;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

}

func (g *cpyGen) genTypeTypeCheck(sym *symbol) {
	g.decl.Printf(
		"\n/* check-type function for %[1]s */\n",
		sym.gofmt(),
	)
	g.decl.Printf("static int\n")
	g.decl.Printf(
		"cpy_func_%[1]s_check(PyObject *self);\n",
		sym.id,
	)

	g.impl.Printf(
		"\n/* check-type function for %[1]s */\n",
		sym.gofmt(),
	)
	g.impl.Printf("static int\n")
	g.impl.Printf(
		"cpy_func_%[1]s_check(PyObject *self) {\n",
		sym.id,
	)
	g.impl.Indent()
	g.impl.Printf(
		"return PyObject_TypeCheck(self, &cpy_type_%sType);\n",
		sym.id,
	)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

}
