// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

// genPatchLeakFuncName outputs the name of either the function or the class
// 'getter' represented by its arguments that need to have a memory
// leak induced by the interaction of pybindgen and gopy.
// These leaks occur because the cgo bindings use C.CString
// to allocate a string on the C heap, as in:
//
//	return C.CString(cstrings.StringValue(C.GoString(s), int(n)))
//
// but the calling pybindgen code then takes a copy of this
// newly allocated string indirectly by calling Py_BuildValue (see below)
// without freeing the copy returned by the function or getter method.
// Arguably pybindgen should provide a means of directly requesting
// this behaviour (simularly to the transfer_ownership options it provides),
// but as of now there is no means of doing so. As a work around the
// output of pybindgen is modified to insert the free in the appropriate
// functions identified by genPatchLeakFuncName.
//
// The patch-leaks.go (see gen.go) file patches these code fragments
// to insert free(retval) immediately after the call to PyBuildValue.
// The patch-leaks.go generated file will perform this patch for
// functions whose names are embedded in it by genPatchLeakFuncName.
//
//  .....
//  retval = cstrings_StringValue(s, n);
//  py_retval = Py_BuildValue((char *) "s", retval);
//  .....
//
// This patching is required for top level functions that return
// strings directly and also for getters that return strings that
// are embedded in maps, structs etc.
// NOTE that it's ok to output the same function name multiple times (eg.
// for a shared structs).
func (g *pyGen) genPatchLeakFuncName(ret *Var, fsym *Func) {
	if ret == nil || fsym == nil || ret.GoType() == nil {
		return
	}
	g.recurse(ret.GoType(), "_wrap__"+fsym.pkg.Name()+"_", fsym.ID())
}

func (g *pyGen) recurse(gotype types.Type, prefix, name string) {
	switch t := gotype.(type) {
	case *types.Basic:
		if t.Kind() == types.String {
			g.leakfile.Printf("    %q,\n", prefix+name)
		}
	case *types.Map:
		g.recurse(t.Elem(), prefix, "Map_"+t.Key().String()+"_string_elem")
	case *types.Named:
		o := t.Obj()
		if o == nil || o.Pkg() == nil {
			return
		}
		g.recurse(t.Underlying(), prefix, o.Pkg().Name()+"_"+o.Name())
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)
			g.recurse(f.Type(), prefix, name+"_"+f.Name()+"_Get")
		}
	case *types.Slice:
		g.recurse(t.Elem(), prefix, "Slice_string_elem")
	}
}

// genFuncSig generates just the signature for binding
// returns false if function is not suitable for python
// binding (e.g., multiple return values)
func (g *pyGen) genFuncSig(sym *symbol, fsym *Func) bool {
	isMethod := (sym != nil)

	if fsym.sig == nil {
		return false
	}

	gname, gdoc, err := extractPythonName(fsym.GoName(), fsym.Doc())
	if err != nil {
		return false
	}
	ifchandle, gdoc := isIfaceHandle(gdoc)

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

	var (
		goArgs []string
		pyArgs []string
		wpArgs []string
	)

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
		if ifchandle && arg.sym.goname == "interface{}" {
			goArgs = append(goArgs, fmt.Sprintf("%s %s", anm, CGoHandle))
			pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '%s')", PyHandle, anm))
		} else {
			goArgs = append(goArgs, fmt.Sprintf("%s %s", anm, sarg.cgoname))
			if sarg.cpyname == "PyObject*" {
				pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '%s', transfer_ownership=False)", sarg.cpyname, anm))
			} else {
				pyArgs = append(pyArgs, fmt.Sprintf("param('%s', '%s')", sarg.cpyname, anm))
			}
		}
		wpArgs = append(wpArgs, anm)
	}

	// support for optional arg to run in a separate go routine -- only if no return val
	if nres == 0 {
		goArgs = append(goArgs, "goRun C.char")
		pyArgs = append(pyArgs, "param('bool', 'goRun')")
		wpArgs = append(wpArgs, "goRun=False")
	}

	switch {
	case isMethod:
		mnm := sym.id + "_" + fsym.GoName()

		g.gofile.Printf("\n//export %s\n", mnm)
		g.gofile.Printf("func %s(", mnm)

		g.pybuild.Printf("mod.add_function('%s', ", mnm)

		g.pywrap.Printf("def %s(", gname)
	default:
		g.gofile.Printf("\n//export %s\n", fsym.ID())
		g.gofile.Printf("func %s(", fsym.ID())

		g.pybuild.Printf("mod.add_function('%s', ", fsym.ID())

		g.pywrap.Printf("def %s(", gname)
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

		g.genPatchLeakFuncName(ret, fsym)

		if sret.cpyname == "PyObject*" {
			g.pybuild.Printf("retval('%s', caller_owns_return=True)", sret.cpyname)
		} else {
			g.pybuild.Printf("retval('%s')", sret.cpyname)
		}
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

func (g *pyGen) genMethod(s *symbol, o *Func) {
	if g.genFuncSig(s, o) {
		g.genFuncBody(s, o)
	}
}

func isIfaceHandle(gdoc string) (bool, string) {
	const PythonIface = "\ngopy:interface=handle"
	if idx := strings.Index(gdoc, PythonIface); idx > 0 {
		gdoc = gdoc[:idx] + gdoc[idx+len(PythonIface)+1:]
		return true, gdoc
	}
	return false, gdoc
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

	_, gdoc, _ := extractPythonName(fsym.GoName(), fsym.Doc())
	ifchandle, gdoc := isIfaceHandle(gdoc)

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
	g.pywrap.Printf(`"""%s"""`, gdoc)
	g.pywrap.Printf("\n")

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
			`vifc, __err := gopyh.VarFromHandleTry((gopyh.CGoHandle)(_handle), "%s")
if __err != nil {
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
		g.gofile.Printf("var __err error\n")
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
		switch {
		case ifchandle && arg.sym.goname == "interface{}":
			na = fmt.Sprintf(`gopyh.VarFromHandle((gopyh.CGoHandle)(%s), "interface{}")`, anm)
		case arg.sym.isSignature():
			na = fmt.Sprintf("%s", arg.sym.py2go)
		case arg.sym.py2go != "":
			na = fmt.Sprintf("%s(%s)%s", arg.sym.py2go, anm, arg.sym.py2goParenEx)
		default:
			na = anm
		}
		callArgs = append(callArgs, na)
		switch {
		case arg.sym.goname == "interface{}":
			if ifchandle {
				wrapArgs = append(wrapArgs, fmt.Sprintf("%s.handle", anm))
			} else {
				wrapArgs = append(wrapArgs, anm)
			}
		case arg.sym.hasHandle():
			wrapArgs = append(wrapArgs, fmt.Sprintf("%s.handle", anm))
		default:
			wrapArgs = append(wrapArgs, anm)
		}
	}

	hasRetCvt := false
	hasAddrOfTmp := false
	if nres > 0 {
		ret := res[0]
		switch {
		case rvIsErr:
			g.gofile.Printf("__err = ")
		case nres == 2:
			g.gofile.Printf("cret, __err := ")
		case ret.sym.hasHandle() && !ret.sym.isPtrOrIface():
			hasAddrOfTmp = true
			g.gofile.Printf("cret := ")
		case ret.sym.go2py != "":
			hasRetCvt = true
			g.gofile.Printf("return %s(", ret.sym.go2py)
		default:
			g.gofile.Printf("return ")
		}
	}
	if nres == 0 {
		wrapArgs = append(wrapArgs, "goRun")
	}
	g.pywrap.Printf("%s)", strings.Join(wrapArgs, ", "))
	if rvHasHandle {
		g.pywrap.Printf(")")
	}

	funCall := ""
	if isMethod {
		if sym.isStruct() {
			funCall = fmt.Sprintf("gopyh.Embed(vifc, reflect.TypeOf(%s{})).(%s).%s(%s)", nonPtrName(symNm), symNm, fsym.GoName(), strings.Join(callArgs, ", "))
		} else {
			funCall = fmt.Sprintf("vifc.(%s).%s(%s)", symNm, fsym.GoName(), strings.Join(callArgs, ", "))
		}
	} else {
		funCall = fmt.Sprintf("%s(%s)", fsym.GoFmt(), strings.Join(callArgs, ", "))
	}
	if hasRetCvt {
		ret := res[0]
		funCall += fmt.Sprintf(")%s", ret.sym.go2pyParenEx)
	}

	if nres == 0 {
		g.gofile.Printf("if boolPyToGo(goRun) {\n")
		g.gofile.Indent()
		g.gofile.Printf("go %s\n", funCall)
		g.gofile.Outdent()
		g.gofile.Printf("} else {\n")
		g.gofile.Indent()
		g.gofile.Printf("%s\n", funCall)
		g.gofile.Outdent()
		g.gofile.Printf("}")
	} else {
		g.gofile.Printf("%s\n", funCall)
	}

	if rvIsErr || nres == 2 {
		g.gofile.Printf("\n")
		g.gofile.Printf("if __err != nil {\n")
		g.gofile.Indent()
		g.gofile.Printf("estr := C.CString(__err.Error())\n")
		g.gofile.Printf("C.PyErr_SetString(C.PyExc_RuntimeError, estr)\n")
		if rvIsErr {
			g.gofile.Printf("return estr\n") // NOTE: leaked string
		} else {
			g.gofile.Printf("C.free(unsafe.Pointer(estr))\n") // python should have converted, safe
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
		if rvIsErr {
			g.gofile.Printf("return C.CString(\"\")") // NOTE: leaked string
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
