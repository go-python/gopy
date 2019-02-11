// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

func (g *pybindGen) genFuncSig(sym *symbol, fsym Func) {
	isMethod := (sym != nil)

	switch {
	case isMethod:
		mnm := sym.goname + "_" + fsym.GoName()

		g.gofile.Printf("\n//export %s\n", mnm)
		g.gofile.Printf("func %s(", mnm)

		g.pyfile.Printf("\n# wrapping %s.%s\n", sym.goname, fsym.GoName())
		g.pyfile.Printf("mod.add_function('%s', ", mnm)

	default:
		g.gofile.Printf("\n//export %s\n", fsym.GoName())
		g.gofile.Printf("func %s(", fsym.GoName())

		g.pyfile.Printf("\n# wrapping %s\n", fsym.GoName())
		g.pyfile.Printf("mod.add_function('%s', ", fsym.GoName())
	}
	sig := fsym.sig
	args := sig.Params()
	res := sig.Results()

	nargs := 0
	nres := 0

	goArgs := []string{}
	pyArgs := []string{}
	if isMethod {
		goArgs = append(goArgs, "_handle *C.char")
		pyArgs = append(pyArgs, "param('char*', '_handle')")
	}

	if args != nil {
		nargs = len(args)
		for i := 0; i < nargs; i++ {
			arg := args[i]
			sarg := g.pkg.syms.symtype(arg.GoType())
			if sarg == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for %q",
					arg.Name(),
				))
			}
			goArgs = append(goArgs, fmt.Sprintf("%s %s", arg.Name(), sarg.cgoname))
			pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '%s')", sarg.cpyname, arg.Name()))
		}
	}

	goRet := ""
	nres = len(res)
	if nres > 0 {
		ret := res[0]
		sret := g.pkg.syms.symtype(ret.GoType())
		if sret == nil {
			panic(fmt.Errorf(
				"gopy: could not find symbol for %q",
				ret.Name(),
			))
		}
		// todo: check for pointer return type
		g.pyfile.Printf("retval('%s')", sret.cpyname)
		goRet = fmt.Sprintf("%s", sret.cgoname)
	} else {
		g.pyfile.Printf("None")
	}

	if len(goArgs) > 0 {
		pstr := strings.Join(pyArgs, ", ")
		g.pyfile.Printf(", [%v])\n", pstr)

		gstr := strings.Join(goArgs, ", ")
		g.gofile.Printf("%v) %v", gstr, goRet)
	} else {
		g.gofile.Printf(") %v", goRet)

		g.pyfile.Printf(", [])\n")
	}
}

func (g *pybindGen) genFunc(o Func) {
	g.genFuncSig(nil, o)
	g.genFuncBody(nil, o)
}

func (g *pybindGen) genMethod(s Struct, o Func) {
	g.genFuncSig(s.sym, o)
	g.genFuncBody(s.sym, o)
}

func (g *pybindGen) genFuncBody(sym *symbol, m Func) {
	isMethod := (sym != nil)

	sig := m.Signature()
	res := sig.Results()
	args := sig.Params()
	nres := len(res)

	if nres > 2 {
		panic(fmt.Errorf("gopy: function/method with more than 2 results not supported! (%s)", m.ID()))
	}
	if nres == 2 && !m.err {
		panic(fmt.Errorf("gopy: function/method with 2 results must have error as second result! (%s)", m.ID()))
	}
	rvIsErr := false // set to true if the main return is an error
	if nres == 1 {
		ret := res[0]
		if isErrorType(ret.GoType()) {
			rvIsErr = true
		}
	}

	g.gofile.Printf(" {\n")
	g.gofile.Indent()
	if isMethod {
		g.gofile.Printf(
			`vifc, err := varHand.varFmHandleTry(_handle, "*%s")
if err != nil {
`, sym.gofmt())
		g.gofile.Indent()
		if nres > 0 {
			ret := res[0]
			if ret.sym.c2py != "" {
				g.gofile.Printf("return %s(%s)\n", ret.sym.c2py, ret.sym.zval)
			} else {
				g.gofile.Printf("return %s\n", ret.sym.zval)
			}
		} else {
			g.gofile.Printf("return\n")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
	} else if rvIsErr {
		g.gofile.Printf("var err error\n")
	}

	callArgs := []string{}
	for _, arg := range args {
		if arg.sym.py2c != "" {
			callArgs = append(callArgs, fmt.Sprintf("%s(%s)", arg.sym.py2c, arg.Name()))
		} else {
			callArgs = append(callArgs, arg.Name())
		}
	}

	hasRetCvt := false
	if nres > 0 {
		ret := res[0]
		if rvIsErr {
			g.gofile.Printf("err = ")
		} else if nres == 2 {
			g.gofile.Printf("cret, err := ")
		} else {
			if ret.sym.c2py != "" {
				g.gofile.Printf("return %s(", ret.sym.c2py)
				hasRetCvt = true
			} else {
				g.gofile.Printf("return ")
			}
		}
	}
	if isMethod {
		g.gofile.Printf("vifc.(*%v).%v(%v)", sym.gofmt(), m.GoName(), strings.Join(callArgs, ", "))
	} else {
		g.gofile.Printf("%v(%v)", m.GoFmt(), strings.Join(callArgs, ", "))
	}
	if hasRetCvt {
		g.gofile.Printf(")")
	} else if rvIsErr || nres == 2 {
		g.gofile.Printf("\n")
		g.gofile.Printf("if err != nil {\n")
		g.gofile.Indent()
		g.gofile.Printf("C.PyErr_SetString(C.PyExc_RuntimeError, C.CString(err.Error()))\n")
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
		if rvIsErr {
			g.gofile.Printf("return C.CString(err.Error())")
		} else {
			ret := res[0]
			if ret.sym.c2py != "" {
				g.gofile.Printf("return %s(cret)", ret.sym.c2py)
			} else {
				g.gofile.Printf("return cret")
			}
		}
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}
