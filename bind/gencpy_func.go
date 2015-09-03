// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

func (g *cpyGen) _genFunc(sym *symbol, fsym *symbol) {
	isMethod := (sym != nil)

	switch {
	case isMethod:
		g.decl.Printf("\n/* wrapping %s.%s */\n", sym.gofmt(), fsym.goname)
		g.decl.Printf("static PyObject*\n")
		g.decl.Printf(
			"cpy_func_%[1]s(%[2]s *self, PyObject *args, PyObject *kwds);\n",
			fsym.id,
			sym.cpyname,
		)

		g.impl.Printf(
			"\n/* wrapping %s.%s */\n",
			sym.gofmt(),
			fsym.goname,
		)
		g.impl.Printf("static PyObject*\n")
		g.impl.Printf(
			"cpy_func_%[1]s(%[2]s *self, PyObject *args, PyObject *kwds) {\n",
			fsym.id,
			sym.cpyname,
		)

	default:
		g.decl.Printf("\n/* wrapping %s */\n", fsym.goname)
		g.decl.Printf("static PyObject*\n")
		g.decl.Printf(
			"cpy_func_%[1]s(PyObject *self, PyObject *args, PyObject *kwds);\n",
			fsym.id,
		)

		g.impl.Printf(
			"\n/* wrapping %s */\n",
			fsym.goname,
		)
		g.impl.Printf("static PyObject*\n")
		g.impl.Printf(
			"cpy_func_%[1]s(PyObject *self, PyObject *args, PyObject *kwds) {\n",
			fsym.id,
		)
	}
	g.impl.Indent()
	sig := fsym.GoType().Underlying().(*types.Signature)
	args := sig.Params()
	res := sig.Results()

	nargs := 0
	nres := 0

	funcArgs := []string{}
	if isMethod {
		funcArgs = append(funcArgs, "self->cgopy")
	}

	if args != nil {
		nargs = args.Len()
		for i := 0; i < nargs; i++ {
			arg := args.At(i)
			sarg := g.pkg.syms.symtype(arg.Type())
			if sarg == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for %q",
					arg.String(),
				))
			}
			g.impl.Printf("%[1]s arg%03d;\n",
				sarg.cgoname,
				i,
			)
			funcArgs = append(funcArgs, fmt.Sprintf("arg%03d", i))
		}
	}

	if res != nil {
		nres = res.Len()
		switch nres {
		case 0:
			// no-op

		case 1:
			ret := res.At(0)
			sret := g.pkg.syms.symtype(ret.Type())
			if sret == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for %q",
					ret.String(),
				))
			}
			g.impl.Printf("%[1]s ret;\n", sret.cgoname)

		default:
			g.impl.Printf(
				"struct cgo_func_%[1]s_return ret;\n",
				fsym.id,
			)
		}
	}

	g.impl.Printf("\n")

	if nargs > 0 {
		g.impl.Printf("if (!PyArg_ParseTuple(args, ")
		format := []string{}
		pyaddrs := []string{}
		for i := 0; i < nargs; i++ {
			sarg := g.pkg.syms.symtype(args.At(i).Type())
			vname := fmt.Sprintf("arg%03d", i)
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

	if nres > 0 {
		g.impl.Printf("ret = ")
	}

	g.impl.Printf("cgo_func_%[1]s(%[2]s);\n\n",
		fsym.id,
		strings.Join(funcArgs, ", "),
	)

	if nres <= 0 {
		g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n\n")
		return
	}

	if hasError(sig) {
		switch nres {
		case 1:
			g.impl.Printf("if (!_cgopy_ErrorIsNil(ret)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(ret);\n")
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
			g.impl.Printf("if (!_cgopy_ErrorIsNil(ret.r1)) {\n")
			g.impl.Indent()
			g.impl.Printf("const char* c_err_str = _cgopy_ErrorString(ret.r1);\n")
			g.impl.Printf("PyErr_SetString(PyExc_RuntimeError, c_err_str);\n")
			g.impl.Printf("free((void*)c_err_str);\n")
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			ret := res.At(0)
			sret := g.pkg.syms.symtype(ret.Type())
			g.impl.Printf("return %s(&ret.r0);\n", sret.c2py)
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			return

		default:
			panic(fmt.Errorf(
				"bind: function/method with more than 2 results not supported! (%s)",
				fsym.id,
			))
		}
	}

	ret := res.At(0)
	sret := g.pkg.syms.symtype(ret.Type())
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

	funcArgs := []string{}

	res := sig.Results()
	args := sig.Params()
	var recv *Var
	if sig.Recv() != nil {
		recv = sig.Recv()
		recv.genRecvDecl(g.impl)
		funcArgs = append(funcArgs, recv.getFuncArg())
	}

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
	g.impl.Printf("cgopy_seq_buffer_t ibuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("cgopy_seq_buffer_t obuf = cgopy_seq_buffer_new();\n")
	g.impl.Printf("\n")

	// fill input seq-buffer
	if len(args) > 0 {
		for i, arg := range args {
			g.genWrite(fmt.Sprintf("arg%03d", i), "ibuf", arg.sym)
		}
	}

	if len(res) > 0 {
		g.impl.Printf("c_gopy_ret = ")
	}
	g.impl.Printf("cgo_func_%[1]s(%[2]s);\n", id, strings.Join(funcArgs, ", "))

	g.impl.Printf("\n")

	if len(res) <= 0 {
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
			g.impl.Printf("return NULL;\n")
			g.impl.Outdent()
			g.impl.Printf("}\n\n")
			g.impl.Printf("Py_INCREF(Py_None);\nreturn Py_None;\n")
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
			if f.ctor {
				ret := res[0]
				g.impl.Printf("PyObject *o = cpy_func_%[1]s_new(&%[2]sType, 0, 0);\n",
					ret.sym.id,
					ret.sym.cpyname,
				)
				g.impl.Printf("if (o == NULL) {\n")
				g.impl.Indent()
				g.impl.Printf("return NULL;\n")
				g.impl.Outdent()
				g.impl.Printf("}\n")
				g.impl.Printf("((%[1]s*)o)->cgopy = c_gopy_ret.r0;\n",
					ret.sym.cpyname,
				)
				g.impl.Printf("return o;\n")
				return
			}
			pyfmt, _ := res[0].getArgBuildValue()
			g.impl.Printf("return Py_BuildValue(%q, c_gopy_ret.r0);\n", pyfmt)
			return

		default:
			panic(fmt.Errorf(
				"bind: function/method with more than 2 results not supported! (%s)",
				f.ID(),
			))
		}
	}

	if f.ctor {
		ret := res[0]
		g.impl.Printf("PyObject *o = cpy_func_%[1]s_new(&%[2]sType, 0, 0);\n",
			ret.sym.id,
			ret.sym.cpyname,
		)
		g.impl.Printf("if (o == NULL) {\n")
		g.impl.Indent()
		g.impl.Printf("return NULL;\n")
		g.impl.Outdent()
		g.impl.Printf("}\n")
		g.impl.Printf("((%[1]s*)o)->cgopy = c_gopy_ret;\n",
			ret.sym.cpyname,
		)
		g.impl.Printf("return o;\n")
		return
	}

	format := []string{}
	funcArgs = []string{}
	switch len(res) {
	case 1:
		ret := res[0]
		ret.name = "gopy_ret"
		pyfmt, pyaddrs := ret.getArgBuildValue()
		format = append(format, pyfmt)
		funcArgs = append(funcArgs, pyaddrs...)
	default:
		for _, ret := range res {
			pyfmt, pyaddrs := ret.getArgBuildValue()
			format = append(format, pyfmt)
			funcArgs = append(funcArgs, pyaddrs...)
		}
	}

	g.impl.Printf("return Py_BuildValue(%q, %s);\n",
		strings.Join(format, ""),
		strings.Join(funcArgs, ", "),
	)
}

func (g *cpyGen) genWrite(valName, seqName string, sym *symbol) {
	if isErrorType(sym.GoType()) {
		g.impl.Printf("cgopy_seq_write_error(%s, %s);\n", seqName, valName)
	}

	switch sym.GoType() {

	}
}
