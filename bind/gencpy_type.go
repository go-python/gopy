// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

func (g *cpyGen) genType(typ Type) {
	if _, ok := typ.GoType().(*types.Basic); ok {
		return
	}
	if _, ok := typ.GoType().(*types.Named); !ok {
		return
	}

	g.decl.Printf("\n/* --- decls for type %v --- */\n\n", typ.gofmt())

	g.decl.Printf("/* Python type for %v\n", typ.gofmt())
	g.decl.Printf(" */\ntypedef struct {\n")
	g.decl.Indent()
	g.decl.Printf("PyObject_HEAD\n")
	g.decl.Printf("%[1]s cgopy; /* handle to %[2]s */\n",
		typ.CGoName(),
		typ.gofmt(),
	)
	g.decl.Printf("gopy_efacefunc eface;\n")
	g.decl.Outdent()
	g.decl.Printf("} %s;\n", typ.CPyName())
	g.decl.Printf("\n\n")

	g.impl.Printf("\n\n/* --- impl for %s */\n\n", typ.gofmt())

	g.genTypeNew(typ)
	g.genTypeDealloc(typ)
	g.genTypeInit(typ)
	g.genTypeMembers(typ)
	g.genTypeMethods(typ)

	g.genTypeProtocols(typ)

	tpAsBuffer := "0"
	tpAsSequence := "0"
	tpFlags := "Py_TPFLAGS_DEFAULT"
	if typ.isArray() || typ.isSlice() {
		tpAsBuffer = fmt.Sprintf("&%[1]s_tp_as_buffer", typ.CPyName())
		tpAsSequence = fmt.Sprintf("&%[1]s_tp_as_sequence", typ.CPyName())
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
	if sig, ok := typ.GoType().Underlying().(*types.Signature); ok {
		if sig.Recv() == nil {
			// only generate tp_call for functions (not methods)
			tpCall = fmt.Sprintf("(ternaryfunc)cpy_func_%[1]s_tp_call", typ.ID())
		}
	}

	g.impl.Printf("static PyTypeObject %sType = {\n", typ.CPyName())
	g.impl.Indent()
	g.impl.Printf("PyObject_HEAD_INIT(NULL)\n")
	g.impl.Printf("0,\t/*ob_size*/\n")
	g.impl.Printf("\"%s\",\t/*tp_name*/\n", typ.gofmt())
	g.impl.Printf("sizeof(%s),\t/*tp_basicsize*/\n", typ.CPyName())
	g.impl.Printf("0,\t/*tp_itemsize*/\n")
	g.impl.Printf("(destructor)cpy_func_%s_dealloc,\t/*tp_dealloc*/\n", typ.ID())
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
	g.impl.Printf("cpy_func_%s_tp_str,\t/*tp_str*/\n", typ.ID())
	g.impl.Printf("0,\t/*tp_getattro*/\n")
	g.impl.Printf("0,\t/*tp_setattro*/\n")
	g.impl.Printf("%s,\t/*tp_as_buffer*/\n", tpAsBuffer)
	g.impl.Printf("%s,\t/*tp_flags*/\n", tpFlags)
	g.impl.Printf("%q,\t/* tp_doc */\n", typ.Doc())
	g.impl.Printf("0,\t/* tp_traverse */\n")
	g.impl.Printf("0,\t/* tp_clear */\n")
	g.impl.Printf("0,\t/* tp_richcompare */\n")
	g.impl.Printf("0,\t/* tp_weaklistoffset */\n")
	g.impl.Printf("0,\t/* tp_iter */\n")
	g.impl.Printf("0,\t/* tp_iternext */\n")
	g.impl.Printf("%s_methods,             /* tp_methods */\n", typ.CPyName())
	g.impl.Printf("0,\t/* tp_members */\n")
	g.impl.Printf("%s_getsets,\t/* tp_getset */\n", typ.CPyName())
	g.impl.Printf("0,\t/* tp_base */\n")
	g.impl.Printf("0,\t/* tp_dict */\n")
	g.impl.Printf("0,\t/* tp_descr_get */\n")
	g.impl.Printf("0,\t/* tp_descr_set */\n")
	g.impl.Printf("0,\t/* tp_dictoffset */\n")
	g.impl.Printf("(initproc)cpy_func_%s_init,      /* tp_init */\n", typ.ID())
	g.impl.Printf("0,                         /* tp_alloc */\n")
	g.impl.Printf("cpy_func_%s_new,\t/* tp_new */\n", typ.ID())
	g.impl.Outdent()
	g.impl.Printf("};\n\n")

	g.genTypeConverter(typ)
	g.genTypeTypeCheck(typ)
}

func (g *cpyGen) genTypeNew(typ Type) {
	f := typ.funcs.new

	g.decl.Printf("\n/* tp_new for %s */\n", typ.gofmt())
	g.decl.Printf(
		"static PyObject*\ncpy_func_%s_new(PyTypeObject *type, PyObject *args, PyObject *kwds);\n",
		typ.ID(),
	)

	g.impl.Printf("\n/* tp_new */\n")
	g.impl.Printf(
		"static PyObject*\ncpy_func_%s_new(PyTypeObject *type, PyObject *args, PyObject *kwds) {\n",
		typ.ID(),
	)
	g.impl.Indent()
	g.impl.Printf("%s *self;\n", typ.CPyName())
	g.impl.Printf("cgopy_seq_buffer ibuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("cgopy_seq_buffer obuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("\n")
	g.impl.Printf("self = (%s *)type->tp_alloc(type, 0);\n", typ.CPyName())

	g.impl.Printf("cgopy_seq_send(%q, %d, ibuf->buf, ibuf->len, &obuf->buf, &obuf->len);\n\n",
		f.Descriptor(),
		uhash(f.ID()),
	)
	g.impl.Printf("self->cgopy = cgopy_seq_buffer_read_int32(obuf);\n")
	//g.impl.Printf("self->eface = (gopy_efacefunc)cgo_func_%s_eface;\n", sym.id)
	g.impl.Printf("return (PyObject*)self;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeDealloc(typ Type) {
	g.decl.Printf("\n/* tp_dealloc for %s */\n", typ.gofmt())
	g.decl.Printf("static void\ncpy_func_%[1]s_dealloc(%[2]s *self);\n",
		typ.ID(),
		typ.CPyName(),
	)

	g.impl.Printf("\n/* tp_dealloc for %s */\n", typ.gofmt())
	g.impl.Printf("static void\ncpy_func_%[1]s_dealloc(%[2]s *self) {\n",
		typ.ID(),
		typ.CPyName(),
	)
	g.impl.Indent()
	g.impl.Printf("cgopy_seq_destroy_ref(self->cgopy);\n")
	g.impl.Printf("self->ob_type->tp_free((PyObject*)self);\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeInit(typ Type) {
	if styp := typ.Struct(); styp != nil {
		g.genStructInit(typ)
		return
	}

	g.decl.Printf("\n/* tp_init for %s */\n", typ.gofmt())
	g.decl.Printf(
		"static int\ncpy_func_%[1]s_init(%[2]s *self, PyObject *args, PyObject *kwds);\n",
		typ.ID(),
		typ.CPyName(),
	)

	g.impl.Printf("\n/* tp_init */\n")
	g.impl.Printf(
		"static int\ncpy_func_%[1]s_init(%[2]s *self, PyObject *args, PyObject *kwds) {\n",
		typ.ID(),
		typ.CPyName(),
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
		typ.GoName(),
		nargs,
	)
	g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
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
	g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	// FIXME(sbinet) handle slices, arrays and funcs
	switch t := typ.GoType().Underlying().(type) {
	case *types.Basic:
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()
		bsym := g.pkg.syms.symtype(t)
		g.impl.Printf(
			"if (!%s(arg, &self->cgopy)) {\n",
			bsym.py2c,
		)
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Outdent()
		g.impl.Printf("}\n\n")

	case *types.Array:
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()

		g.impl.Printf("if (!PySequence_Check(arg)) {\n")
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
		g.impl.Printf("\"%s.__init__ takes a sequence as argument\");\n", typ.GoName())
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		esym := g.pkg.syms.symtype(t.Elem())
		if esym == nil {
			panic(fmt.Errorf(
				"gopy: could not find symbol for element of %q",
				typ.gofmt(),
			))
		}

		g.impl.Printf("Py_ssize_t len = PySequence_Size(arg);\n")
		g.impl.Printf("if (len == -1) {\n")
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("if (len > %d) {\n", t.Len())
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_ValueError, ")
		g.impl.Printf("\"%s.__init__ takes a sequence of size at most %d\");\n",
			typ.GoName(),
			t.Len(),
		)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("Py_ssize_t i = 0;\n")
		g.impl.Printf("for (i = 0; i < len; i++) {\n")
		g.impl.Indent()
		g.impl.Printf("PyObject *elt = PySequence_GetItem(arg, i);\n")
		g.impl.Printf("if (cpy_func_%[1]s_ass_item(self, i, elt)) {\n", typ.ID())
		g.impl.Indent()
		g.impl.Printf("Py_XDECREF(elt);\n")
		g.impl.Printf(
			"PyErr_SetString(PyExc_TypeError, \"invalid type (expected a %s)\");\n",
			esym.goname,
		)
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("Py_XDECREF(elt);\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n") // for-loop

		g.impl.Outdent()
		g.impl.Printf("}\n\n") // if-arg

	case *types.Slice:
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()

		g.impl.Printf("if (!PySequence_Check(arg)) {\n")
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
		g.impl.Printf("\"%s.__init__ takes a sequence as argument\");\n", typ.ID())
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.impl.Printf("if (!cpy_func_%[1]s_inplace_concat(self, arg)) {\n", typ.ID())
		g.impl.Indent()
		g.impl.Printf("goto cpy_label_%s_init_fail;\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n") // if-arg

	case *types.Map:
		g.impl.Printf("if (arg != NULL) {\n")
		g.impl.Indent()
		//g.impl.Printf("//put map __init__ functions here")
		g.impl.Outdent()
		g.impl.Printf("}\n\n") // if-arg

	case *types.Signature:
		//TODO(sbinet)

	case *types.Interface:
		//TODO(sbinet): check the argument implements the interface.

	default:
		panic(fmt.Errorf(
			"gopy: tp_init for %s not handled",
			typ.gofmt(),
		))
	}

	g.impl.Printf("return 0;\n")
	g.impl.Outdent()

	g.impl.Printf("\ncpy_label_%s_init_fail:\n", typ.ID())
	g.impl.Indent()
	g.impl.Printf("Py_XDECREF(arg);\n")
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeMembers(typ Type) {
	if typ.isStruct() {
		g.genStructMembers(typ)
		return
	}

	g.decl.Printf("\n/* tp_getset for %s */\n", typ.gofmt())
	g.impl.Printf("\n/* tp_getset for %s */\n", typ.gofmt())
	g.impl.Printf("static PyGetSetDef %s_getsets[] = {\n", typ.CPyName())
	g.impl.Indent()
	g.impl.Printf("{NULL} /* Sentinel */\n")
	g.impl.Outdent()
	g.impl.Printf("};\n\n")
}

func (g *cpyGen) genTypeMethods(typ Type) {
	g.decl.Printf("\n/* methods for %s */\n", typ.gofmt())
	for _, meth := range typ.meths {
		g._genFunc(meth)
	}
	g.impl.Printf("\n/* methods for %s */\n", typ.gofmt())
	g.impl.Printf("static PyMethodDef %s_methods[] = {\n", typ.CPyName())
	g.impl.Indent()
	for _, m := range typ.meths {
		margs := "METH_VARARGS"
		sig := m.Signature()
		if sig.Params() == nil || sig.Params().Len() <= 0 {
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

func (g *cpyGen) genTypeProtocols(typ Type) {
	g.genTypeTPStr(typ)
	if typ.isSlice() || typ.isArray() {
		g.genTypeTPAsSequence(typ)
		g.genTypeTPAsBuffer(typ)
	}
	if typ.isSignature() {
		g.genTypeTPCall(typ)
	}
}

func (g *cpyGen) genTypeTPStr(typ Type) {
	f := typ.funcs.str
	g.decl.Printf("\n/* __str__ support for %[1]s.%[2]s */\n",
		f.Package().Name(),
		typ.GoName(),
	)
	g.decl.Printf(
		"static PyObject*\ncpy_func_%s_tp_str(PyObject *self);\n",
		typ.ID(),
	)

	g.impl.Printf(
		"static PyObject*\ncpy_func_%s_tp_str(PyObject *self) {\n",
		typ.ID(),
	)

	g.impl.Indent()
	if f != (Func{}) {
		g.genFuncBody(f)
	} else {
		g.impl.Printf("PyObject *pystr = NULL;\n")
		g.impl.Printf("cgopy_seq_bytearray str;\n")
		g.impl.Printf("\n")
		g.impl.Printf("cgopy_seq_buffer ibuf = cgopy_seq_buffer_new();\n")
		g.impl.Printf("cgopy_seq_buffer obuf = cgopy_seq_buffer_new();\n")
		g.impl.Printf("\n")

		g.impl.Printf("int32_t c_self = ((%[1]s*)self)->cgopy;\n", typ.CPyName())
		g.impl.Printf("pystr = cgopy_cnv_c2py_string(&str);\n")
		g.impl.Printf("\n")
		g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
		g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
		g.impl.Printf("\n")
		g.impl.Printf("return pystr;\n")
	}
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeTPAsSequence(typ Type) {
	g.decl.Printf("\n/* sequence support for %s */\n", typ.gofmt())

	var arrlen int64
	var etyp types.Type
	switch t := typ.GoType().(type) {
	case *types.Array:
		etyp = t.Elem()
		arrlen = t.Len()
	case *types.Slice:
		etyp = t.Elem()
	case *types.Named:
		switch tn := t.Underlying().(type) {
		case *types.Array:
			etyp = tn.Elem()
			arrlen = tn.Len()
		case *types.Slice:
			etyp = tn.Elem()
		default:
			panic(fmt.Errorf(
				"gopy: unhandled type [%#v] (go=%s)",
				typ,
				typ.gofmt(),
			))
		}
	default:
		panic(fmt.Errorf(
			"gopy: unhandled type [%#v] (go=%s)",
			typ,
			typ.gofmt(),
		))
	}
	esym := g.pkg.syms.symtype(etyp)
	if esym == nil {
		panic(fmt.Errorf("gopy: could not retrieve element type of %v",
			typ.gofmt(),
		))
	}

	switch g.lang {
	case 2:

		g.decl.Printf("\n/* len */\n")
		g.decl.Printf("static Py_ssize_t\ncpy_func_%[1]s_len(%[2]s *self);\n",
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* len */\n")
		g.impl.Printf("static Py_ssize_t\ncpy_func_%[1]s_len(%[2]s *self) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		if typ.isArray() {
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
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* item */\n")
		g.impl.Printf("static PyObject*\n")
		g.impl.Printf("cpy_func_%[1]s_item(%[2]s *self, Py_ssize_t i) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		g.impl.Printf("PyObject *pyitem = NULL;\n")
		if typ.isArray() {
			g.impl.Printf("if (i < 0 || i >= %d) {\n", arrlen)
		} else {
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (i < 0 || i >= slice->len) {\n")
		}
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_IndexError, ")
		g.impl.Printf("\"array index out of range\");\n")
		g.impl.Printf("return NULL;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("%[1]s item = cgo_func_%[2]s_item(self->cgopy, i);\n",
			esym.cgoname,
			typ.CPyName(),
		)
		g.impl.Printf("pyitem = %[1]s(&item);\n", esym.c2py)
		g.impl.Printf("return pyitem;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* ass_item */\n")
		g.decl.Printf("static int\n")
		g.decl.Printf("cpy_func_%[1]s_ass_item(%[2]s *self, Py_ssize_t i, PyObject *v);\n",
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* ass_item */\n")
		g.impl.Printf("static int\n")
		g.impl.Printf("cpy_func_%[1]s_ass_item(%[2]s *self, Py_ssize_t i, PyObject *v) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		g.impl.Printf("%[1]s c_v;\n", esym.cgoname)
		if typ.isArray() {
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
		g.impl.Printf("cgo_func_%[1]s_ass_item(self->cgopy, i, c_v);\n", typ.ID())
		g.impl.Printf("return 0;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		sq_inplace_concat := "0"
		// append
		if typ.isSlice() {
			sq_inplace_concat = fmt.Sprintf(
				"cpy_func_%[1]s_inplace_concat",
				typ.ID(),
			)

			g.decl.Printf("\n/* append-item */\n")
			g.decl.Printf("static int\n")
			g.decl.Printf("cpy_func_%[1]s_append(%[2]s *self, PyObject *v);\n",
				typ.ID(),
				typ.CPyName(),
			)

			g.impl.Printf("\n/* append-item */\n")
			g.impl.Printf("static int\n")
			g.impl.Printf("cpy_func_%[1]s_append(%[2]s *self, PyObject *v) {\n",
				typ.ID(),
				typ.CPyName(),
			)
			g.impl.Indent()
			g.impl.Printf("%[1]s c_v;\n", esym.cgoname)
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (v == NULL) { return 0; }\n") // FIXME(sbinet): semantics?
			g.impl.Printf("if (!%[1]s(v, &c_v)) { return -1; }\n", esym.py2c)
			g.impl.Printf("cgo_func_%[1]s_append(self->cgopy, c_v);\n", typ.ID())
			g.impl.Printf("return 0;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.decl.Printf("\n/* inplace-concat */\n")
			g.decl.Printf("static PyObject*\n")
			g.decl.Printf("cpy_func_%[1]s_inplace_concat(%[2]s *self, PyObject *v);\n",
				typ.ID(),
				typ.CPyName(),
			)

			g.impl.Printf("\n/* inplace-item */\n")
			g.impl.Printf("static PyObject*\n")
			g.impl.Printf("cpy_func_%[1]s_inplace_concat(%[2]s *self, PyObject *v) {\n",
				typ.ID(),
				typ.CPyName(),
			)
			g.impl.Indent()
			// FIXME(sbinet) do the append in one go?
			g.impl.Printf("if (!PySequence_Check(v)) {\n")
			g.impl.Indent()
			g.impl.Printf("PyErr_SetString(PyExc_TypeError, ")
			g.impl.Printf("\"%s.__iadd__ takes a sequence as argument\");\n", typ.GoName())
			g.impl.Printf("goto cpy_label_%s_inplace_concat_fail;\n", typ.ID())
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.impl.Printf("Py_ssize_t len = PySequence_Size(v);\n")
			g.impl.Printf("if (len == -1) {\n")
			g.impl.Indent()
			g.impl.Printf("goto cpy_label_%s_inplace_concat_fail;\n", typ.ID())
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

			g.impl.Printf("Py_ssize_t i = 0;\n")
			g.impl.Printf("for (i = 0; i < len; i++) {\n")
			g.impl.Indent()
			g.impl.Printf("PyObject *elt = PySequence_GetItem(v, i);\n")
			g.impl.Printf("if (cpy_func_%[1]s_append(self, elt)) {\n", typ.ID())
			g.impl.Indent()
			g.impl.Printf("Py_XDECREF(elt);\n")
			g.impl.Printf(
				"PyErr_Format(PyExc_TypeError, \"invalid type (got=%%s, expected a %s)\", Py_TYPE(elt)->tp_name);\n",
				esym.goname,
			)
			g.impl.Printf("goto cpy_label_%s_inplace_concat_fail;\n", typ.ID())
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			g.impl.Printf("Py_XDECREF(elt);\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n") // for-loop

			g.impl.Printf("return (PyObject*)self;\n")
			g.impl.Outdent()

			g.impl.Printf("\ncpy_label_%s_inplace_concat_fail:\n", typ.ID())
			g.impl.Indent()
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")

		}

		g.impl.Printf("\n/* tp_as_sequence */\n")
		g.impl.Printf("static PySequenceMethods %[1]s_tp_as_sequence = {\n", typ.CPyName())
		g.impl.Indent()
		g.impl.Printf("(lenfunc)cpy_func_%[1]s_len,\n", typ.ID())
		g.impl.Printf("(binaryfunc)0,\n")   // array_concat,               sq_concat
		g.impl.Printf("(ssizeargfunc)0,\n") //array_repeat,                 /*sq_repeat
		g.impl.Printf("(ssizeargfunc)cpy_func_%[1]s_item,\n", typ.ID())
		g.impl.Printf("(ssizessizeargfunc)0,\n") // array_slice,             /*sq_slice
		g.impl.Printf("(ssizeobjargproc)cpy_func_%[1]s_ass_item,\n", typ.ID())
		g.impl.Printf("(ssizessizeobjargproc)0,\n") //array_ass_slice,      /*sq_ass_slice
		g.impl.Printf("(objobjproc)0,\n")           //array_contains,                 /*sq_contains
		g.impl.Printf("(binaryfunc)%s,\n", sq_inplace_concat)
		g.impl.Printf("(ssizeargfunc)0\n") //array_inplace_repeat          /*sq_inplace_repeat
		g.impl.Outdent()
		g.impl.Printf("};\n\n")

	case 3:
	}
}

func (g *cpyGen) genTypeTPAsBuffer(typ Type) {
	g.decl.Printf("\n/* buffer support for %s */\n", typ.gofmt())

	g.decl.Printf("\n/* __get_buffer__ impl for %s */\n", typ.gofmt())
	g.decl.Printf("static int\n")
	g.decl.Printf(
		"cpy_func_%[1]s_getbuffer(PyObject *self, Py_buffer *view, int flags);\n",
		typ.ID(),
	)

	var esize int64
	var arrlen int64
	esym := g.pkg.syms.symtype(typ.GoType())
	switch o := typ.GoType().Underlying().(type) {
	case *types.Array:
		esize = g.pkg.sz.Sizeof(o.Elem())
		arrlen = o.Len()
	case *types.Slice:
		esize = g.pkg.sz.Sizeof(o.Elem())
	default:
		panic(fmt.Errorf("type %v is not a sequence", typ.gofmt()))
	}

	g.impl.Printf("\n/* __get_buffer__ impl for %s */\n", typ.gofmt())
	g.impl.Printf("static int\n")
	g.impl.Printf(
		"cpy_func_%[1]s_getbuffer(PyObject *self, Py_buffer *view, int flags) {\n",
		typ.ID(),
	)
	g.impl.Indent()
	g.impl.Printf("if (view == NULL) {\n")
	g.impl.Indent()
	g.impl.Printf("PyErr_SetString(PyExc_ValueError, ")
	g.impl.Printf("\"NULL view in getbuffer\");\n")
	g.impl.Printf("return -1;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
	g.impl.Printf("%[1]s *py = (%[1]s*)self;\n", typ.CPyName())
	switch {
	case typ.isArray():
		g.impl.Printf("void *array = (void*)(py->cgopy);\n")
		g.impl.Printf("view->obj = (PyObject*)py;\n")
		g.impl.Printf("view->buf = (void*)array;\n")
		g.impl.Printf("view->len = %d;\n", arrlen)
		g.impl.Printf("view->readonly = 0;\n")
		g.impl.Printf("view->itemsize = %d;\n", esize)
		g.impl.Printf("view->format = %q;\n", esym.pybuf)
		g.impl.Printf("view->ndim = 1;\n")
		g.impl.Printf("view->shape = (Py_ssize_t*)&view->len;\n")
	case typ.isSlice():
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
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* readbuffer */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf(
			"cpy_func_%[1]s_readbuffer(%[2]s *self, Py_ssize_t index, const void **ptr) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		g.impl.Printf("if (index != 0) {\n")
		g.impl.Indent()
		g.impl.Printf("PyErr_SetString(PyExc_SystemError, ")
		g.impl.Printf("\"Accessing non-existent array segment\");\n")
		g.impl.Printf("return -1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		switch {
		case typ.isArray():
			g.impl.Printf("*ptr = (void*)self->cgopy;\n")
			g.impl.Printf("return %d;\n", arrlen)
		case typ.isSlice():
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
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* writebuffer */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf(
			"cpy_func_%[1]s_writebuffer(%[2]s *self, Py_ssize_t segment, void **ptr) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		g.impl.Printf("return cpy_func_%[1]s_readbuffer(self, segment, (const void**)ptr);\n",
			typ.ID(),
		)
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* segcount */\n")
		g.decl.Printf("static Py_ssize_t\n")
		g.decl.Printf("cpy_func_%[1]s_segcount(%[2]s *self, Py_ssize_t *lenp);\n",
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* segcount */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf("cpy_func_%[1]s_segcount(%[2]s *self, Py_ssize_t *lenp) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		switch {
		case typ.isArray():
			g.impl.Printf("if (lenp) { *lenp = %d; }\n", arrlen)
		case typ.isSlice():
			g.impl.Printf("GoSlice *slice = (GoSlice*)(self->cgopy);\n")
			g.impl.Printf("if (lenp) { *lenp = slice->len; }\n")
		}
		g.impl.Printf("return 1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

		g.decl.Printf("\n/* charbuffer */\n")
		g.decl.Printf("static Py_ssize_t\n")
		g.decl.Printf("cpy_func_%[1]s_charbuffer(%[2]s *self, Py_ssize_t segment, const char **ptr);\n",
			typ.ID(),
			typ.CPyName(),
		)

		g.impl.Printf("\n/* charbuffer */\n")
		g.impl.Printf("static Py_ssize_t\n")
		g.impl.Printf("cpy_func_%[1]s_charbuffer(%[2]s *self, Py_ssize_t segment, const char **ptr) {\n",
			typ.ID(),
			typ.CPyName(),
		)
		g.impl.Indent()
		g.impl.Printf("return cpy_func_%[1]s_readbuffer(self, segment, (const void**)ptr);\n", typ.ID())
		g.impl.Outdent()
		g.impl.Printf("}\n\n")

	case 3:
		// no-op
	}

	switch g.lang {
	case 2:
		g.impl.Printf("\n/* tp_as_buffer */\n")
		g.impl.Printf("static PyBufferProcs %[1]s_tp_as_buffer = {\n", typ.CPyName())
		g.impl.Indent()
		g.impl.Printf("(readbufferproc)cpy_func_%[1]s_readbuffer,\n", typ.ID())
		g.impl.Printf("(writebufferproc)cpy_func_%[1]s_writebuffer,\n", typ.ID())
		g.impl.Printf("(segcountproc)cpy_func_%[1]s_segcount,\n", typ.ID())
		g.impl.Printf("(charbufferproc)cpy_func_%[1]s_charbuffer,\n", typ.ID())
		g.impl.Printf("(getbufferproc)cpy_func_%[1]s_getbuffer,\n", typ.ID())
		g.impl.Printf("(releasebufferproc)0,\n")
		g.impl.Outdent()
		g.impl.Printf("};\n\n")
	case 3:

		g.impl.Printf("\n/* tp_as_buffer */\n")
		g.impl.Printf("static PyBufferProcs %[1]s_tp_as_buffer = {\n", typ.CPyName())
		g.impl.Indent()
		g.impl.Printf("(getbufferproc)cpy_func_%[1]s_getbuffer,\n", typ.ID())
		g.impl.Printf("(releasebufferproc)0,\n")
		g.impl.Outdent()
		g.impl.Printf("};\n\n")
	}
}

func (g *cpyGen) genTypeTPCall(typ Type) {
	if !typ.isSignature() {
		return
	}

	sig := typ.GoType().Underlying().(*types.Signature)
	if sig.Recv() != nil {
		// don't generate tp_call for methods.
		return
	}

	g.decl.Printf("\n/* tp_call */\n")
	g.decl.Printf("static PyObject *\n")
	g.decl.Printf(
		"cpy_func_%[1]s_tp_call(%[2]s *self, PyObject *args, PyObject *other);\n",
		typ.ID(),
		typ.CPyName(),
	)

	g.impl.Printf("\n/* tp_call */\n")
	g.impl.Printf("static PyObject *\n")
	g.impl.Printf(
		"cpy_func_%[1]s_tp_call(%[2]s *self, PyObject *args, PyObject *other) {\n",
		typ.ID(),
		typ.CPyName(),
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
				typ.ID(),
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
	g.impl.Printf("cgo_func_%[1]s_call(%[2]s);\n", typ.ID(), strings.Join(funcArgs, ", "))

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
				typ.ID(),
			))
		}
	}

	ret := res[0]
	sret := g.pkg.syms.symtype(ret.GoType())
	g.impl.Printf("return %s(&c_gopy_ret);\n", sret.c2py)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genTypeConverter(typ Type) {
	g.decl.Printf("\n/* converters for %s - %s */\n",
		typ.ID(),
		typ.GoName(),
	)
	g.decl.Printf("static int\n")
	g.decl.Printf("cgopy_cnv_py2c_%[1]s(PyObject *o, int32_t *addr);\n",
		typ.ID(),
	)
	g.decl.Printf("static PyObject*\n")
	g.decl.Printf("cgopy_cnv_c2py_%[1]s(int32_t *addr);\n\n",
		typ.ID(),
	)

	g.impl.Printf("static int\n")
	g.impl.Printf("cgopy_cnv_py2c_%[1]s(PyObject *o, int32_t *addr) {\n",
		typ.ID(),
	)
	g.impl.Indent()
	g.impl.Printf("%s *self = NULL;\n", typ.CPyName())
	if typ.isInterface() {
		g.impl.Printf("if (%s) {\n", fmt.Sprintf(typ.CPyCheck(), "o"))
		g.impl.Indent()
		g.impl.Printf("self = (%s *)o;\n", typ.CPyName())
		g.impl.Printf("*addr = self->cgopy;\n")
		g.impl.Printf("return 1;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		g.impl.Printf("GoInterface *iface = (GoInterface*)(addr);\n")
		g.impl.Printf("*iface = ((gopy_object*)o)->eface((gopy_object*)o);\n")
		g.impl.Printf("return 1;\n")
	} else {
		g.impl.Printf("self = (%s *)o;\n", typ.CPyName())
		g.impl.Printf("*addr = self->cgopy;\n")
		g.impl.Printf("return 1;\n")
	}
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

	g.impl.Printf("static PyObject*\n")
	g.impl.Printf("cgopy_cnv_c2py_%[1]s(int32_t *addr) {\n", typ.ID())
	g.impl.Indent()
	g.impl.Printf("PyObject *o = cpy_func_%[1]s_new(&%[2]sType, 0, 0);\n",
		typ.ID(),
		typ.CPyName(),
	)
	g.impl.Printf("if (o == NULL) {\n")
	g.impl.Indent()
	g.impl.Printf("return NULL;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n")
	g.impl.Printf("((%[1]s*)o)->cgopy = *addr;\n", typ.CPyName())
	g.impl.Printf("return o;\n")
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

}

func (g *cpyGen) genTypeTypeCheck(typ Type) {
	g.decl.Printf(
		"\n/* check-type function for %[1]s */\n",
		typ.gofmt(),
	)
	g.decl.Printf("static int\n")
	g.decl.Printf(
		"cpy_func_%[1]s_check(PyObject *self);\n",
		typ.ID(),
	)

	g.impl.Printf(
		"\n/* check-type function for %[1]s */\n",
		typ.gofmt(),
	)
	g.impl.Printf("static int\n")
	g.impl.Printf(
		"cpy_func_%[1]s_check(PyObject *self) {\n",
		typ.ID(),
	)
	g.impl.Indent()
	g.impl.Printf(
		"return PyObject_TypeCheck(self, &cpy_type_%sType);\n",
		typ.ID(),
	)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")

}
