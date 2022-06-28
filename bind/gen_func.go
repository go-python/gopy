// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"log"
	"strconv"
	"strings"
)

func buildPyTuple(fsym *Func) bool {
	npyres := len(fsym.sig.Results())
	if fsym.haserr {
		if !NoPyExceptions {
			npyres -= 1
		}
	}

	return (npyres > 1)
}

func (g *pyGen) recurse(gotype types.Type, prefix, name string) {
	switch t := gotype.(type) {
	case *types.Basic:
		// pass
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

	gname := fsym.GoName()
	if g.cfg.RenameCase {
		gname = toSnakeCase(gname)
	}

	gname, gdoc, err := extractPythonName(gname, fsym.Doc())
	if err != nil {
		return false
	}
	ifchandle, gdoc := isIfaceHandle(gdoc)

	sig := fsym.sig
	args := sig.Params()
	res := sig.Results()
	nargs := 0
	nres := len(res)
	npyres := nres
	if fsym.haserr {
		if !NoPyExceptions {
			npyres -= 1
		}
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

		if i != nargs-1 || !fsym.isVariadic {
			wpArgs = append(wpArgs, anm)
		}
	}

	// support for optional arg to run in a separate go routine -- only if no return val
	if nres == 0 {
		goArgs = append(goArgs, "goRun C.char")
		pyArgs = append(pyArgs, "param('bool', 'goRun')")
		wpArgs = append(wpArgs, "goRun=False")
	}

	// To support variadic args, we add *args at the end.
	if fsym.isVariadic {
		wpArgs = append(wpArgs, "*args")
	}

	// When building the pybindgen builder code, we start with
	// a function that adds function calls with exception checking.
	// But given specific return types, we may want to add more
	// behavior to the wrapped function code gen.
	retvalsToFree := make([]string, 0, npyres)
	if npyres > 0 {
		for i := 0; i < npyres; i++ {
			switch t := res[i].GoType().(type) {
			case *types.Basic:
				// string return types need special memory leak patches
				// to free the allocated char*
				if t.Kind() == types.String {
					retvalsToFree = append(retvalsToFree, strconv.Itoa(i))
				}
			}
		}
	}
	pyTupleBuilt := "True"
	if !buildPyTuple(fsym) {
		pyTupleBuilt = "False"
	}
	addFuncName := "add_checked_function_generator(" + pyTupleBuilt + ", [" + strings.Join(retvalsToFree, ", ") + "])"

	switch {
	case isMethod:
		mnm := sym.id + "_" + fsym.GoName()

		g.gofile.Printf("\n//export %s\n", mnm)
		g.gofile.Printf("func %s(", mnm)

		g.pybuild.Printf("%s(mod, '%s', ", addFuncName, mnm)

		g.pywrap.Printf("def %s(", gname)
	default:
		g.gofile.Printf("\n//export %s\n", fsym.ID())
		g.gofile.Printf("func %s(", fsym.ID())

		g.pybuild.Printf("%s(mod, '%s', ", addFuncName, fsym.ID())

		g.pywrap.Printf("def %s(", gname)
	}

	goRet := ""
	if npyres == 0 {
		g.pybuild.Printf("None")
	} else if buildPyTuple(fsym) {
		// We are returning PyTuple*. Setup pybindgen accordingly.
		g.pybuild.Printf("retval('PyObject*', caller_owns_return=True)")

		// On Go side, return *C.PyObject.
		goRet = "unsafe.Pointer"
	} else {
		ownership := ""
		pyrets := make([]string, npyres, npyres)
		gorets := make([]string, npyres, npyres)
		for i := 0; i < npyres; i++ {
			sret := current.symtype(res[i].GoType())
			if sret == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for %q",
					res[i].Name(),
				))
			}
			gorets[i] = sret.cgoname
			pyrets[i] = "'" + sret.cpyname + "'"
			if sret.cpyname == "PyObject*" {
				ownership = ", caller_owns_return=True"
			}
		}

		g.pybuild.Printf("retval(%s%s)", strings.Join(pyrets, ", "), ownership)

		goRet = strings.Join(gorets, ", ")
		if npyres > 1 {
			goRet = "(" + goRet + ")"
		}
	}

	if len(goArgs) > 0 {
		g.gofile.Printf("%v) %v", strings.Join(goArgs, ", "), goRet)

		g.pybuild.Printf(", [%v])\n", strings.Join(pyArgs, ", "))

		g.pywrap.Printf("%v)", strings.Join(wpArgs, ", "))

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
	const PythonIface = "gopy:interface=handle"
	if idx := strings.Index(gdoc, PythonIface); idx >= 0 {
		gdoc = gdoc[:idx] + gdoc[idx+len(PythonIface)+1:]
		return true, gdoc
	}
	return false, gdoc
}

func isPointer(pyfmt string) bool {
	if pyfmt == "s" {
		return true
	}
	return false
}

func (g *pyGen) generateReturn(buildPyTuple bool, npyres int, results []*Var, retvals *[]string) {
	g.gofile.Printf("\n")
	valueCalls := make([]string, npyres, npyres)
	for i := 0; i < npyres; i++ {
		result := results[i]
		sret := current.symtype(result.GoType())
		if sret == nil {
			panic(fmt.Errorf(
				"gopy: could not find symbol for %q",
				result.Name(),
			))
		}
		formatStr := sret.pyfmt
		if sret.pyfmt == "" {
			formatStr = "?"
		}

		retval := ""
		if retvals != nil {
			retval = (*retvals)[i]
		} else if result.sym.zval != "" {
			retval = result.sym.zval
		} else {
			fmt.Printf("gopy: programmer error: empty zval zero value in symbol: %v\n", result.sym)
		}

		if result.sym.go2py != "" {
			retval = result.sym.go2py + "(" + retval + ")" + result.sym.go2pyParenEx
		}

		if buildPyTuple {
			buildValueFunc := "C.Py_BuildValue1"
			typeCast := "unsafe.Pointer"
			if !isPointer(formatStr) {
				buildValueFunc = "C.Py_BuildValue2"
				typeCast = "C.longlong"
				formatStr = "L"
			}
			valueCalls[i] = fmt.Sprintf("%s(C.CString(\"%s\"), %s(%s))",
				buildValueFunc,
				formatStr,
				typeCast,
				retval)
		} else {
			valueCalls[i] = retval
		}
	}

	if npyres == 0 {
		g.gofile.Printf("return\n")
	} else if buildPyTuple {
		g.gofile.Printf("retTuple := C.PyTuple_New(%d);\n", npyres)
		for i := 0; i < npyres; i++ {
			g.gofile.Printf("C.PyTuple_SetItem(retTuple, %d, %s);\n", i, valueCalls[i])
		}
		g.gofile.Printf("return unsafe.Pointer(retTuple);")
	} else {
		g.gofile.Printf("return %s\n", strings.Join(valueCalls, ", "))
	}
}

func (g *pyGen) genFuncBody(sym *symbol, fsym *Func) {
	isMethod := (sym != nil)
	isIface := false
	symNm := ""
	if isMethod {
		symNm = sym.goname
		isIface = sym.isInterface()
		if !isIface {
			symNm = "*" + symNm
		}
	}

	pkgname := g.cfg.Name

	sig := fsym.Signature()
	res := sig.Results()
	args := sig.Params()
	nres := len(res)
	npyres := nres
	rvHasErr := false // set to true if the main return is an error
	if fsym.haserr {
		if NoPyExceptions {
			rvHasErr = true
		} else {
			npyres -= 1
		}
	}

	_, gdoc, _ := extractPythonName(fsym.GoName(), fsym.Doc())
	ifchandle, gdoc := isIfaceHandle(gdoc)

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
		g.generateReturn(buildPyTuple(fsym), npyres, res, nil)
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
	} else if rvHasErr {
		g.gofile.Printf("var __err error\n")
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
		if i == len(args)-1 && fsym.isVariadic {
			na = na + "..."
		}
		callArgs = append(callArgs, na)
		switch {
		case arg.sym.goname == "interface{}":
			if ifchandle {
				wrapArgs = append(wrapArgs, fmt.Sprintf("(-1 if %s == None else %s.handle)", anm, anm))
			} else {
				wrapArgs = append(wrapArgs, anm)
			}
		case arg.sym.hasHandle():
			wrapArgs = append(wrapArgs, fmt.Sprintf("(-1 if %s == None else %s.handle)", anm, anm))
		default:
			wrapArgs = append(wrapArgs, anm)
		}

		// To support variadic args, we add *args at the end.
		if fsym.isVariadic && i == len(args)-1 {
			packagePrefix := ""
			if arg.sym.gopkg.Name() != fsym.pkg.Name() {
				packagePrefix = arg.sym.gopkg.Name() + "."
			}
			g.pywrap.Printf("%s = %s%s(args)\n", anm, packagePrefix, arg.sym.id)
		}
	}

	// pywrap output
	mnm := fsym.ID()
	if isMethod {
		mnm = sym.id + "_" + fsym.GoName()
	}

	log.Println(mnm)

	if nres == 0 {
		wrapArgs = append(wrapArgs, "goRun")
	}

	pyFuncCall := fmt.Sprintf("_%s.%s(%s)", pkgname, mnm, strings.Join(wrapArgs, ", "))
	if npyres > 0 {
		retvars := make([]string, nres, nres)
		for i := 0; i < npyres; i++ {
			retvars[i] = "ret" + strconv.Itoa(i)
			if res[i].sym.hasHandle() {
				retvars[i] = "_" + retvars[i]
			}
		}

		if fsym.haserr {
			// Only used if npyres == nres.
			retvars[nres-1] = "__err"
		}

		// May drop the error var, if it is not supposed to be returned.
		retvars = retvars[0:npyres]

		// Call upstream method and collect returns.
		g.pywrap.Printf(fmt.Sprintf("%s = %s\n", strings.Join(retvars, ", "), pyFuncCall))

		// ReMap handle returns from pyFuncCall.
		for i := 0; i < npyres; i++ {
			if res[i].sym.hasHandle() {
				cvnm := res[i].sym.pyPkgId(g.pkg.pkg)
				g.pywrap.Printf("ret%d = %s(handle=_ret%d)\n", i, cvnm, i)
				retvars[i] = "ret" + strconv.Itoa(i)
			}
		}

		g.pywrap.Printf("return %s", strings.Join(retvars, ", "))
	} else {
		g.pywrap.Printf(pyFuncCall)
	}

	g.pywrap.Printf("\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	goFuncCall := ""
	if isMethod {
		if sym.isStruct() {
			goFuncCall = fmt.Sprintf("gopyh.Embed(vifc, reflect.TypeOf(%s{})).(%s).%s(%s)",
				nonPtrName(symNm),
				symNm,
				fsym.GoName(),
				strings.Join(callArgs, ", "))
		} else {
			goFuncCall = fmt.Sprintf("vifc.(%s).%s(%s)",
				symNm,
				fsym.GoName(),
				strings.Join(callArgs, ", "))
		}
	} else {
		goFuncCall = fmt.Sprintf("%s(%s)",
			fsym.GoFmt(),
			strings.Join(callArgs, ", "))
	}

	if nres > 0 {
		retvals := make([]string, nres, nres)
		for i := 0; i < npyres; i++ {
			retvals[i] = "cret" + strconv.Itoa(i)
		}
		if fsym.haserr {
			retvals[nres-1] = "__err_ret"
		}

		// Call upstream method and collect returns.
		g.gofile.Printf(fmt.Sprintf("%s := %s\n", strings.Join(retvals, ", "), goFuncCall))

		// ReMap handle returns from pyFuncCall.
		for i := 0; i < npyres; i++ {
			if res[i].sym.hasHandle() && !res[i].sym.isPtrOrIface() {
				retvals[i] = "&" + retvals[i]
			}
		}

		if fsym.haserr {
			g.gofile.Printf("\n")

			if rvHasErr {
				retvals[npyres-1] = "estr" // NOTE: leaked string
				g.gofile.Printf("var estr C.CString\n")
				g.gofile.Printf("if __err_ret != nil {\n")
				g.gofile.Indent()
				g.gofile.Printf("estr = C.CString(__err_ret.Error())// NOTE: leaked string\n") // NOTE: leaked string
				g.gofile.Outdent()
				g.gofile.Printf("} else {\n")
				g.gofile.Indent()
				g.gofile.Printf("estr = C.CString(\"\")// NOTE: leaked string\n") // NOTE: leaked string
				g.gofile.Outdent()
				g.gofile.Printf("}\n")
			} else {
				g.gofile.Printf("if __err_ret != nil {\n")
				g.gofile.Indent()
				g.gofile.Printf("estr := C.CString(__err_ret.Error())\n") // NOTE: freed string
				g.gofile.Printf("C.PyErr_SetString(C.PyExc_RuntimeError, estr)\n")
				g.gofile.Printf("C.free(unsafe.Pointer(estr))\n") // python should have converted, safe
				g.gofile.Outdent()
				g.gofile.Printf("}\n")
			}
		}

		g.generateReturn(buildPyTuple(fsym), npyres, res, &retvals)
	} else {
		g.gofile.Printf("if boolPyToGo(goRun) {\n")
		g.gofile.Indent()
		g.gofile.Printf("go %s\n", goFuncCall)
		g.gofile.Outdent()
		g.gofile.Printf("} else {\n")
		g.gofile.Indent()
		g.gofile.Printf("%s\n", goFuncCall)
		g.gofile.Outdent()
		g.gofile.Printf("}")
	}

	g.gofile.Printf("\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
}
