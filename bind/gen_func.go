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

		g.pybuild.Printf("\n# wrapping %s.%s\n", sym.goname, fsym.GoName())
		g.pybuild.Printf("mod.add_function('%s', ", mnm)

		g.pywrap.Printf("def %s(", fsym.GoName())

	default:
		g.gofile.Printf("\n//export %s\n", fsym.GoName())
		g.gofile.Printf("func %s(", fsym.GoName())

		g.pybuild.Printf("\n# wrapping %s\n", fsym.GoName())
		g.pybuild.Printf("mod.add_function('%s', ", fsym.GoName())

		g.pywrap.Printf("def %s(", fsym.GoName())

	}

	sig := fsym.sig
	args := sig.Params()
	res := sig.Results()

	nargs := 0
	nres := 0

	goArgs := []string{}
	pyArgs := []string{}
	wpArgs := []string{}
	if isMethod {
		goArgs = append(goArgs, "_handle *C.char")
		pyArgs = append(pyArgs, "param('char*', '_handle')")
		wpArgs = append(wpArgs, "self")
	}

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
		wpArgs = append(wpArgs, arg.Name())
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
		g.pybuild.Printf("retval('%s')", sret.cpyname)
		goRet = fmt.Sprintf("%s", sret.cgoname)
	} else {
		g.pybuild.Printf("None")
	}

	if len(goArgs) > 0 {
		gstr := strings.Join(goArgs, ", ")
		g.gofile.Printf("%v) %v", gstr, goRet)

		pstr := strings.Join(pyArgs, ", ")
		g.pybuild.Printf(", [%v])\n", pstr)

		wstr := strings.Join(wpArgs, ", ")
		g.pywrap.Printf("%v)", wstr)

	} else {
		g.gofile.Printf(") %v", goRet)

		g.pybuild.Printf(", [])\n")

		g.pywrap.Printf(")")
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

func (g *pybindGen) genFuncBody(sym *symbol, fsym Func) {
	isMethod := (sym != nil)

	pkgname := fsym.Package().Name()

	sig := fsym.Signature()
	res := sig.Results()
	args := sig.Params()
	nres := len(res)

	if nres > 2 {
		panic(fmt.Errorf("gopy: function/method with more than 2 results not supported! (%s)", fsym.ID()))
	}
	if nres == 2 && !fsym.err {
		panic(fmt.Errorf("gopy: function/method with 2 results must have error as second result! (%s)", fsym.ID()))
	}
	rvIsErr := false // set to true if the main return is an error
	if nres == 1 {
		ret := res[0]
		if isErrorType(ret.GoType()) {
			rvIsErr = true
		}
	}

	g.pywrap.Printf(":\n")
	g.pywrap.Indent()

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

	// pywrap output
	mnm := fsym.GoName()
	if isMethod {
		mnm = sym.goname + "_" + mnm
	}
	rvIsPtr := false
	if nres > 0 {
		ret := res[0]
		if ret.sym.isPointer() {
			npnm := ret.sym.goname[1:] // non-pointer name
			rvIsPtr = true
			g.pywrap.Printf("return %s(_%s.%s(", npnm, pkgname, mnm)
		} else {
			g.pywrap.Printf("return _%s.%s(", pkgname, mnm)
		}
	} else {
		g.pywrap.Printf("_%s.%s(", pkgname, mnm)
	}

	callArgs := []string{}
	wrapArgs := []string{}
	if isMethod {
		wrapArgs = append(wrapArgs, "self.handle")
	}
	for _, arg := range args {
		if arg.sym.py2c != "" {
			callArgs = append(callArgs, fmt.Sprintf("%s(%s)", arg.sym.py2c, arg.Name()))
		} else {
			callArgs = append(callArgs, arg.Name())
		}
		if arg.sym.isPointer() {
			wrapArgs = append(wrapArgs, fmt.Sprintf("%s.handle", arg.Name()))
		} else {
			wrapArgs = append(wrapArgs, arg.Name())
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
	g.pywrap.Printf("%s)", strings.Join(wrapArgs, ", "))
	if rvIsPtr {
		g.pywrap.Printf(")")
	}

	if isMethod {
		g.gofile.Printf("vifc.(*%s).%s(%s)", sym.gofmt(), fsym.GoName(), strings.Join(callArgs, ", "))
	} else {
		g.gofile.Printf("%s(%s)", fsym.GoFmt(), strings.Join(callArgs, ", "))
	}
	if hasRetCvt {
		g.gofile.Printf(")")
	} else if rvIsErr || nres == 2 {
		g.gofile.Printf("\n")
		g.gofile.Printf("if err != nil {\n")
		g.gofile.Indent()
		g.gofile.Printf("C.PyErr_SetString(C.PyExc_RuntimeError, C.CString(err.Error()))\n")
		if rvIsErr {
			g.gofile.Printf("return C.CString(err.Error())")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
		if rvIsErr {
			g.gofile.Printf("return C.CString(\"\")")
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

	g.pywrap.Printf("\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")
}
