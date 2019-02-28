// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

// genFuncSig generates just the signature for binding
// returns false if function is not suitable for python
// binding (e.g., multiple return values)
func (g *pyGen) genFuncSig(sym *symbol, fsym *Func) bool {
	isMethod := (sym != nil)

	if fsym.sig == nil {
		return false
	}

	sig := fsym.sig
	args := sig.Params()
	res := sig.Results()
	nargs := 0
	nres := len(res)

	// note: this is enforced in creation of Func, in newFuncFrom
	if nres > 2 {
		return false
	}
	if nres == 2 && !fsym.err {
		return false
	}

	goArgs := []string{}
	pyArgs := []string{}
	wpArgs := []string{}
	if isMethod {
		goArgs = append(goArgs, "_handle CGoHandle")
		pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '_handle')", PyHandle))
		wpArgs = append(wpArgs, "self")
	}

	nargs = len(args)
	for i := 0; i < nargs; i++ {
		arg := args[i]
		sarg := current.symtype(arg.GoType())
		if sarg == nil {
			return false
		}
		anm := pySafeArg(arg.Name(), i)
		goArgs = append(goArgs, fmt.Sprintf("%s %s", anm, sarg.cgoname))
		if sarg.cpyname == "PyObject*" {
			pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '%s', transfer_ownership=False)", sarg.cpyname, anm))
		} else {
			pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '%s')", sarg.cpyname, anm))
		}
		wpArgs = append(wpArgs, anm)
	}

	switch {
	case isMethod:
		mnm := sym.id + "_" + fsym.GoName()

		g.gofile.Printf("\n//export %s\n", mnm)
		g.gofile.Printf("func %s(", mnm)

		g.pybuild.Printf("mod.add_function('%s', ", mnm)

		g.pywrap.Printf("def %s(", fsym.GoName())

	default:
		g.gofile.Printf("\n//export %s\n", fsym.ID())
		g.gofile.Printf("func %s(", fsym.ID())

		g.pybuild.Printf("mod.add_function('%s', ", fsym.ID())

		g.pywrap.Printf("def %s(", fsym.GoName())

	}

	goRet := ""
	nres = len(res)
	if nres > 0 {
		ret := res[0]
		sret := current.symtype(ret.GoType())
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
	return true
}

func (g *pyGen) genFunc(o *Func) {
	if g.genFuncSig(nil, o) {
		g.genFuncBody(nil, o)
	}
}

func (g *pyGen) genMethod(s *Struct, o *Func) {
	if g.genFuncSig(s.sym, o) {
		g.genFuncBody(s.sym, o)
	}
}

func (g *pyGen) genIfcMethod(ifc *Interface, o *Func) {
	if g.genFuncSig(ifc.sym, o) {
		g.genFuncBody(ifc.sym, o)
	}
}

func (g *pyGen) genFuncBody(sym *symbol, fsym *Func) {
	isMethod := (sym != nil)
	isIface := false
	symNm := ""
	if isMethod {
		symNm = sym.gofmt()
		isIface = sym.isInterface()
		if !isIface {
			symNm = "*" + symNm
		}
	}

	pkgname := g.outname

	sig := fsym.Signature()
	res := sig.Results()
	args := sig.Params()
	nres := len(res)

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
	if fsym.hasfun {
		for i, arg := range args {
			if arg.sym.isSignature() {
				g.gofile.Printf("_fun_arg := %s\n", pySafeArg(arg.Name(), i))
			}
		}
	}
	if isMethod {
		g.gofile.Printf(
			`vifc, err := gopyh.VarFmHandleTry((gopyh.CGoHandle)(_handle), "%s")
if err != nil {
`, symNm)
		g.gofile.Indent()
		if nres > 0 {
			ret := res[0]
			if ret.sym.zval == "" {
				fmt.Printf("gopy: programmer error: empty zval zero value in symbol: %v\n", ret.sym)
			}
			if ret.sym.go2py != "" {
				g.gofile.Printf("return %s(%s)%s\n", ret.sym.go2py, ret.sym.zval, ret.sym.go2pyParenEx)
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
	mnm := fsym.ID()
	if isMethod {
		mnm = sym.id + "_" + fsym.GoName()
	}
	rvHasHandle := false
	if nres > 0 {
		ret := res[0]
		if !rvIsErr && ret.sym.hasHandle() {
			rvHasHandle = true
			cvnm := ret.sym.pyPkgId(g.pkg.pkg)
			g.pywrap.Printf("return %s(handle=_%s.%s(", cvnm, pkgname, mnm)
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
	for i, arg := range args {
		na := ""
		anm := pySafeArg(arg.Name(), i)
		if arg.sym.isSignature() {
			na = fmt.Sprintf("%s", arg.sym.py2go)
		} else {
			if arg.sym.py2go != "" {
				na = fmt.Sprintf("%s(%s)%s", arg.sym.py2go, anm, arg.sym.py2goParenEx)
			} else {
				na = anm
			}
		}
		callArgs = append(callArgs, na)
		if arg.sym.hasHandle() {
			wrapArgs = append(wrapArgs, fmt.Sprintf("%s.handle", anm))
		} else {
			wrapArgs = append(wrapArgs, anm)
		}
	}

	hasRetCvt := false
	hasAddrOfTmp := false
	if nres > 0 {
		ret := res[0]
		if rvIsErr {
			g.gofile.Printf("err = ")
		} else if nres == 2 {
			g.gofile.Printf("cret, err := ")
		} else if ret.sym.hasHandle() && !ret.sym.isPtrOrIface() {
			hasAddrOfTmp = true
			g.gofile.Printf("cret := ")
		} else {
			if ret.sym.go2py != "" {
				hasRetCvt = true
				g.gofile.Printf("return %s(", ret.sym.go2py)
			} else {
				g.gofile.Printf("return ")
			}
		}
	}
	g.pywrap.Printf("%s)", strings.Join(wrapArgs, ", "))
	if rvHasHandle {
		g.pywrap.Printf(")")
	}

	if isMethod {
		if sym.isStruct() {
			g.gofile.Printf("gopyh.Embed(vifc, reflect.TypeOf(%s{})).(%s).%s(%s)", nonPtrName(symNm), symNm, fsym.GoName(), strings.Join(callArgs, ", "))
		} else {
			g.gofile.Printf("vifc.(%s).%s(%s)", symNm, fsym.GoName(), strings.Join(callArgs, ", "))
		}
	} else {
		g.gofile.Printf("%s(%s)", fsym.GoFmt(), strings.Join(callArgs, ", "))
	}
	if hasRetCvt {
		ret := res[0]
		g.gofile.Printf(")%s", ret.sym.go2pyParenEx)
	}
	if rvIsErr || nres == 2 {
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
			if ret.sym.go2py != "" {
				if ret.sym.hasHandle() && !ret.sym.isPtrOrIface() {
					g.gofile.Printf("return %s(&cret)%s", ret.sym.go2py, ret.sym.go2pyParenEx)
				} else {
					g.gofile.Printf("return %s(cret)%s", ret.sym.go2py, ret.sym.go2pyParenEx)
				}
			} else {
				g.gofile.Printf("return cret")
			}
		}
	} else if hasAddrOfTmp {
		ret := res[0]
		g.gofile.Printf("\nreturn %s(&cret)%s", ret.sym.go2py, ret.sym.go2pyParenEx)
	}
	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")

	g.pywrap.Printf("\n")
	g.pywrap.Outdent()
}
