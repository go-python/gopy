// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"log"
	"strings"
)

func (g *cpyGen) _genFunc(f Func) {
	recv := f.Signature().Recv()
	isMethod := (recv != nil)

	switch {
	case isMethod:
		g.decl.Printf("\n/* wrapping %s.%s */\n", f.gofmt(), f.GoName())
		g.decl.Printf("static PyObject*\n")
		g.decl.Printf(
			"cpy_func_%[1]s(%[2]s *self, PyObject *args, PyObject *kwds);\n",
			f.ID(),
			f.CPyName(),
		)

		g.impl.Printf(
			"\n/* wrapping %s.%s */\n",
			f.gofmt(),
			f.GoName(),
		)
		g.impl.Printf("static PyObject*\n")
		g.impl.Printf(
			"cpy_func_%[1]s(%[2]s *self, PyObject *args, PyObject *kwds) {\n",
			f.ID(),
			f.CPyName(),
		)

	default:
		g.decl.Printf("\n/* wrapping %s */\n", f.GoName())
		g.decl.Printf("static PyObject*\n")
		g.decl.Printf(
			"cpy_func_%[1]s(PyObject *self, PyObject *args, PyObject *kwds);\n",
			f.ID(),
		)

		g.impl.Printf(
			"\n/* wrapping %s */\n",
			f.GoName(),
		)
		g.impl.Printf("static PyObject*\n")
		g.impl.Printf(
			"cpy_func_%[1]s(PyObject *self, PyObject *args, PyObject *kwds) {\n",
			f.ID(),
		)
	}
	g.impl.Indent()

	sig := f.Signature()
	args := sig.Params()
	res := sig.Results()

	nargs := 0
	nres := 0

	if args != nil {
		nargs = args.Len()
		for i := 0; i < nargs; i++ {
			arg := args.At(0)
			sarg := g.pkg.syms.symtype(arg.Type())
			if sarg == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for %v",
					arg,
				))
			}
			g.impl.Printf("%[1]s _arg%03d;\n",
				sarg.cgoname,
				i,
			)
		}
	}

	if res != nil {
		nres = res.Len()
	}

	switch nres {
	case 0:
		// no-op

	case 1:
		ret := res.At(0)
		sret := g.pkg.syms.symtype(ret.Type())
		if sret == nil {
			panic(fmt.Errorf(
				"gopy: could not find symbol for %v",
				ret,
			))
		}
		g.impl.Printf("%[1]s ret;\n", sret.cgoname)

	default:
		g.impl.Printf(
			"struct cgo_func_%[1]s_return ret;\n",
			f.ID(),
		)
	}

	g.impl.Printf("\n")

	if nargs > 0 {
		g.impl.Printf("if (!PyArg_ParseTuple(args, ")
		format := []string{}
		pyaddrs := []string{}
		for i := 0; i < nargs; i++ {
			sarg := g.pkg.syms.symtype(args.At(i).Type())
			vname := fmt.Sprintf("_arg%03d", i)
			pyfmt, addr := sarg.getArgParse(vname)
			format = append(format, pyfmt)
			pyaddrs = append(pyaddrs, addr...)
		}
		g.impl.Printf("%q, %s)) {\n", strings.Join(format, ""), strings.Join(pyaddrs, ", "))
		g.impl.Indent()
		g.impl.Printf("return NULL;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
	}

	/*
		if nargs > 0 {
			for i := 0; i < nargs; i++ {
				arg := args.At(i)
				sarg := g.pkg.syms.symtype(arg.Type())
				sarg.genFuncPreamble(g.impl)
			}
			g.impl.Printf("\n")
		}
	*/

	// create in/out seq-buffers
	g.impl.Printf("cgopy_seq_buffer ibuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("cgopy_seq_buffer obuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("\n")

	// fill input seq-buffer
	if isMethod {
		g.genWrite("self->cgopy", "ibuf", recv.Type())
	}

	for i := 0; i < nargs; i++ {
		sarg := g.pkg.syms.symtype(args.At(i).Type())
		g.genWrite(fmt.Sprintf("_arg%03d", i), "ibuf", sarg.GoType())
	}

	g.impl.Printf("cgopy_seq_send(%q, %d, ibuf->buf, ibuf->len, &obuf->buf,	&obuf->len);\n\n",
		f.Descriptor(),
		uhash(f.ID()),
	)

	if nres <= 0 {
		g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
		g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
		g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		return
	}

	if hasError(f.Signature()) {
		switch nres {
		case 1:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(ret)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(ret);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			return

		case 2:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(ret.r1)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(ret.r1);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			ret := res.At(0)
			sret := g.pkg.syms.symtype(ret.Type())
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("return %s(&ret.r0);\n", sret.c2py)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			return

		default:
			panic(fmt.Errorf(
				"bind: function/method with more than 2 results not supported! (%s)",
				f.ID(),
			))
		}
	}

	ret := res.At(0)
	sret := g.pkg.syms.symtype(ret.Type())
	g.genRead("ret", "obuf", sret.GoType())

	g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
	g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
	g.impl.Printf("return %s(&ret);\n", sret.c2py)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genFunc(o Func) {

	g.impl.Printf(`
/* pythonization of: %[1]s.%[2]s */
static PyObject*
cpy_func_%[3]s(PyObject *self, PyObject *args) {
`,
		g.pkg.pkg.Name(),
		o.GoName(),
		o.ID(),
	)

	g.impl.Indent()
	g.genFuncBody(o)
	g.impl.Outdent()
	g.impl.Printf("}\n\n")
}

func (g *cpyGen) genFuncBody(f Func) {
	id := f.ID()
	sig := f.Signature()

	res := newVarsFrom(g.pkg, sig.Results()) // FIXME(sbinet) this assumes same g.pkg for all!
	args := newVarsFrom(g.pkg, sig.Params()) // FIXME(sbinet) this assumes same g.pkg for all!
	var recv *Var
	if v := sig.Recv(); v != nil {
		recv = newVarFrom(g.pkg, v) // FIXME(sbinet) this assumes same g.pkg for all!
		recv.genRecvDecl(g.impl)
	}

	for _, arg := range args {
		arg.genDecl(g.impl)
	}

	if len(res) > 0 {
		g.impl.Printf("PyObject *pyout = NULL;\n")
		switch len(res) {
		case 1:
			ret := res[0]
			ret.genRetDecl(g.impl)
		default:
			g.impl.Printf("struct cgo_func_%[1]s_return c_gopy_ret;\n", id)
		}
	}

	g.impl.Printf("\n")

	if recv != nil {
		recv.genRecvImpl(g.impl)
	}

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

	// create in/out seq-buffers
	g.impl.Printf("cgopy_seq_buffer ibuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("cgopy_seq_buffer obuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("\n")

	if recv != nil {
		g.genWrite(fmt.Sprintf("c_%s", recv.Name()), "ibuf", recv.GoType())
	}

	// fill input seq-buffer
	if len(args) > 0 {
		for _, arg := range args {
			g.genWrite(fmt.Sprintf("c_%s", arg.Name()), "ibuf", arg.GoType())
		}
	}

	g.impl.Printf("cgopy_seq_send(%q, %d, ibuf->buf, ibuf->len, &obuf->buf, &obuf->len);\n\n",
		f.Descriptor(),
		uhash(f.ID()),
	)

	if len(res) <= 0 {
		g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
		g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
		g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
		return
	}

	if f.err {
		switch len(res) {
		case 1:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(c_gopy_ret)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(c_gopy_ret);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
			return

		case 2:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(c_gopy_ret.r1)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(c_gopy_ret.r1);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			if f.ctor {
				ret := typeFrom(res[0].GoType())
				g.impl.Printf("PyObject *o = cpy_func_%[1]s_new(&%[2]sType, 0, 0);\n",
					ret.ID(),
					ret.CPyName(),
				)
				g.impl.Printf("if (o == NULL) {\n")
				g.impl.Indent()
				g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
				g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
				g.impl.Printf("return NULL;\n")
				g.impl.Outdent()
				g.impl.Printf("}\n")
				g.impl.Printf("((%[1]s*)o)->cgopy = c_gopy_ret.r0;\n",
					ret.CPyName(),
				)
				g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
				g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
				g.impl.Printf("return o;\n")
				return
			}
			pyfmt, _ := res[0].getArgBuildValue()
			g.impl.Printf("pyout = Py_BuildValue(%q, c_gopy_ret.r0);\n", pyfmt)
			g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
			g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
			g.impl.Printf("return pyout;\n")
			return

		default:
			panic(fmt.Errorf(
				"bind: function/method with more than 2 results not supported! (%s)",
				f.ID(),
			))
		}
	}

	if f.ctor {
		ret := typeFrom(res[0].GoType())
		g.impl.Printf("PyObject *o = cpy_func_%[1]s_new(&%[2]sType, 0, 0);\n",
			ret.ID(),
			ret.CPyName(),
		)
		g.impl.Printf("if (o == NULL) {\n")
		g.impl.Indent()
		g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
		g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
		g.impl.Printf("return NULL;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n")
		g.impl.Printf("((%[1]s*)o)->cgopy = c_gopy_ret;\n",
			ret.CPyName(),
		)
		g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
		g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
		g.impl.Printf("return o;\n")
		return
	}

	format := []string{}
	funcArgs := []string{}
	switch len(res) {
	case 1:
		ret := res[0]
		ret.name = "gopy_ret"
		pyfmt, pyaddrs := ret.getArgBuildValue()
		format = append(format, pyfmt)
		funcArgs = append(funcArgs, pyaddrs...)
		g.genRead("c_gopy_ret", "obuf", ret.GoType())

	default:
		for _, ret := range res {
			pyfmt, pyaddrs := ret.getArgBuildValue()
			format = append(format, pyfmt)
			funcArgs = append(funcArgs, pyaddrs...)
		}
	}

	g.impl.Printf("pyout = Py_BuildValue(%q, %s);\n",
		strings.Join(format, ""),
		strings.Join(funcArgs, ", "),
	)
	g.impl.Printf("cgopy_seq_buffer_free(ibuf);\n")
	g.impl.Printf("cgopy_seq_buffer_free(obuf);\n")
	g.impl.Printf("return pyout;\n")
}

func (g *cpyGen) genWrite(valName, seqName string, T types.Type) {
	if isErrorType(T) {
		g.impl.Printf("cgopy_seq_write_error(%s, %s);\n", seqName, valName)
	}

	switch T := T.(type) {
	case *types.Basic:
		switch T.Kind() {
		case types.Bool:
			log.Fatalf("unhandled type [bool]")
		case types.Int8:
			g.impl.Printf("cgopy_seq_buffer_write_int8(%s, %s);\n", seqName, valName)
		case types.Int16:
			g.impl.Printf("cgopy_seq_buffer_write_int16(%s, %s);\n", seqName, valName)
		case types.Int32:
			g.impl.Printf("cgopy_seq_buffer_write_int32(%s, %s);\n", seqName, valName)
		case types.Int, types.Int64:
			g.impl.Printf("cgopy_seq_buffer_write_int64(%s, %s);\n", seqName, valName)
		case types.Uint8:
			g.impl.Printf("cgopy_seq_buffer_write_uint8(%s, %s);\n", seqName, valName)
		case types.Uint16:
			g.impl.Printf("cgopy_seq_buffer_write_uint16(%s, %s);\n", seqName, valName)
		case types.Uint32:
			g.impl.Printf("cgopy_seq_buffer_write_uint32(%s, %s);\n", seqName, valName)
		case types.Uint, types.Uint64:
			g.impl.Printf("cgopy_seq_buffer_write_uint64(%s, %s);\n", seqName, valName)
		case types.Float32:
			g.impl.Printf("cgopy_seq_buffer_write_float32(%s, %s);\n", seqName, valName)
		case types.Float64:
			g.impl.Printf("cgopy_seq_buffer_write_float64(%s, %s);\n", seqName, valName)
		case types.String:
			g.impl.Printf("cgopy_seq_buffer_write_string(%s, %s);\n", seqName, valName)
		}
	case *types.Named:
		switch u := T.Underlying().(type) {
		case *types.Interface, *types.Pointer, *types.Struct,
			*types.Array, *types.Slice:
			g.impl.Printf("cgopy_seq_buffer_write_int32(%[1]s, %[2]s);\n", seqName, valName)
		case *types.Basic:
			g.genWrite(valName, seqName, u)
		default:
			panic(fmt.Errorf("unsupported, direct named type %s: %s", T, u))
		}
	default:
		g.impl.Printf("/* not implemented %#T */\n", T)
	}
}

func (g *cpyGen) genRead(valName, seqName string, T types.Type) {
	if isErrorType(T) {
		g.impl.Printf("cgopy_seq_read_error(%s, %s);\n", seqName, valName)
	}

	switch T := T.(type) {
	case *types.Basic:
		switch T.Kind() {
		case types.Bool:
			log.Fatalf("unhandled type [bool]")
		case types.Int8:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_int8(%[1]s);\n", seqName, valName)
		case types.Int16:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_int16(%[1]s);\n", seqName, valName)
		case types.Int32:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_int32(%[1]s);\n", seqName, valName)
		case types.Int, types.Int64:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_int64(%[1]s);\n", seqName, valName)
		case types.Uint8:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_uint8(%[1]s);\n", seqName, valName)
		case types.Uint16:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_uint16(%[1]s);\n", seqName, valName)
		case types.Uint32:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_uint32(%[1]s);\n", seqName, valName)
		case types.Uint, types.Uint64:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_uint64(%[1]s);\n", seqName, valName)
		case types.Float32:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_float32(%[1]s);\n", seqName, valName)
		case types.Float64:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_float64(%[1]s);\n", seqName, valName)
		case types.String:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_string(%[1]s);\n", seqName, valName)
		}
	case *types.Named:
		switch u := T.Underlying().(type) {
		case *types.Interface, *types.Pointer, *types.Struct,
			*types.Array, *types.Slice:
			g.impl.Printf("%[2]s = cgopy_seq_buffer_read_int32(%[1]s);\n", seqName, valName)
		case *types.Basic:
			g.genRead(valName, seqName, u)
		default:
			panic(fmt.Errorf("unsupported, direct named type %s: %s", T, u))
		}
	default:
		g.impl.Printf("/* not implemented %#T */\n", T)
	}
}
