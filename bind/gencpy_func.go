// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

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
