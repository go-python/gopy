// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

func (g *cffiGen) genFunc(o Func) {
	sig := o.Signature()
	args := sig.Params()

	var funcArgs []string
	for _, arg := range args {
		funcArgs = append(funcArgs, arg.Name())
	}
	g.wrapper.Printf(`
# pythonization of: %[1]s.%[2]s 
def %[2]s(%[3]s):
    """%[4]s"""
`,
		g.pkg.pkg.Name(),
		o.GoName(),
		strings.Join(funcArgs, ", "),
		o.Doc(),
	)

	g.wrapper.Indent()
	g.genFuncBody(o)
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

func (g *cffiGen) genGetFunc(o Func) {
	sig := o.Signature()
	args := sig.Params()

	var funcArgs []string
	for _, arg := range args {
		funcArgs = append(funcArgs, arg.Name())
	}
	g.wrapper.Printf(`
# pythonization of: %[1]s.%[2]s 
def Get%[2]s(%[3]s):
    """%[4]s"""
`,
		g.pkg.pkg.Name(),
		o.GoName(),
		strings.Join(funcArgs, ", "),
		o.Doc(),
	)

	g.wrapper.Indent()
	g.genFuncBody(o)
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

func (g *cffiGen) genSetFunc(o Func) {
	sig := o.Signature()
	args := sig.Params()

	var funcArgs []string
	for _, arg := range args {
		funcArgs = append(funcArgs, arg.Name())
	}
	g.wrapper.Printf(`
# pythonization of: %[1]s.%[2]s 
def Set%[2]s(%[3]s):
    """%[4]s"""
`,
		g.pkg.pkg.Name(),
		o.GoName(),
		strings.Join(funcArgs, ", "),
		o.Doc(),
	)

	g.wrapper.Indent()
	g.genFuncBody(o)
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

func (g *cffiGen) genFuncBody(f Func) {
	sig := f.Signature()
	res := sig.Results()
	args := sig.Params()
	nres := 0

	var funcArgs []string
	for _, arg := range args {
		if arg.sym.hasConverter() {
			g.wrapper.Printf("%[1]s = _cffi_helper.cffi_%[2]s(%[3]s)\n", arg.getFuncArg(), arg.sym.py2c, arg.Name())
			funcArgs = append(funcArgs, arg.getFuncArg())
		} else {
			funcArgs = append(funcArgs, arg.Name())
		}
	}

	if res != nil {
		nres = len(res)
		if nres > 0 {
			g.wrapper.Printf("cret = ")
		}
	}

	g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s(%[2]s)\n", f.id, strings.Join(funcArgs, ", "))

	if f.err {
		switch nres {
		case 1:
			g.wrapper.Printf("if not _cffi_helper.lib._cgopy_ErrorIsNil(cret):\n")
			g.wrapper.Indent()
			g.wrapper.Printf("c_err_str = _cffi_helper.lib._cgopy_ErrorString(cret)\n")
			g.wrapper.Printf("py_err_str = ffi.string(c_err_str)\n")
			g.wrapper.Printf("_cffi_helper.lib._cgopy_FreeCString(c_err_str)\n")
			g.wrapper.Printf("raise RuntimeError(py_err_str)\n")
			g.wrapper.Outdent()
			g.wrapper.Printf("return\n")
			return
		case 2:
			g.wrapper.Printf("if not _cffi_helper.lib._cgopy_ErrorIsNil(cret.r1):\n")
			g.wrapper.Indent()
			g.wrapper.Printf("c_err_str = _cffi_helper.lib._cgopy_ErrorString(cret.r1)\n")
			g.wrapper.Printf("py_err_str = ffi.string(c_err_str)\n")
			g.wrapper.Printf("_cffi_helper.lib._cgopy_FreeCString(c_err_str)\n")
			g.wrapper.Printf("raise RuntimeError(py_err_str)\n")
			g.wrapper.Outdent()
			if res[0].sym.hasConverter() {
				g.wrapper.Printf("r0 = _cffi_helper.cffi_%[1]s(cret.r0)\n", res[0].sym.c2py)
				g.wrapper.Printf("return r0\n")
			} else {
				g.wrapper.Printf("return cret.r0\n")
			}
			return
		default:
			panic(fmt.Errorf("bind: function/method with more than 2 results not supported! (%s)", f.ID()))
		}
	}

	switch nres {
	case 0:
		// no-op
	case 1:
		ret := res[0]
		if ret.sym.hasConverter() {
			g.wrapper.Printf("ret = _cffi_helper.cffi_%[1]s(cret)\n", ret.sym.c2py)
			g.wrapper.Printf("return ret\n")
		} else {
			g.wrapper.Printf("return cret\n")
		}
	default:
		panic(fmt.Errorf("gopy: Not yet implemeted for multiple return."))
	}
}
