// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
)

// shimABI is what the generated cgo shim differs in, between a backend
// whose shim calls the CPython C API itself (cpythonShim: pybindgen, capi)
// and one whose shim must avoid it (plainCShim, see plainc.go: cffi, pybind11,
// nanobind).  The generators ask it, rather than testing which kind of shim
// is in use.
type shimABI interface {
	// genGoPreamble writes the start of the cgo shim; pkgimport is the
	// import lines of the bound packages.
	genGoPreamble(g *pyGen, pkgimport string)
	// finishGo completes the cgo shim, once the rest of it is written.
	finishGo(g *pyGen)
	// pyGenerate returns the build.py code that writes the module.
	pyGenerate(g *pyGen) string

	// funcSupported reports whether a function with parameters args and
	// results res can be bound; genFuncSig skips it if not.
	funcSupported(g *pyGen, args, res []*Var) bool
	// sliceSupported reports whether a slice or array of esym can be
	// bound; genSliceInit skips it if not.
	sliceSupported(esym *symbol) bool
	// genComplexFunc generates fsym if it needs a path of its own for its
	// complex64/128 values, and reports whether it did.
	genComplexFunc(g *pyGen, fsym *Func) bool
	// complexParts returns the cgo type of each part, if a value of sym
	// crosses as two floats rather than as sym.cgoname.
	complexParts(sym *symbol) (part string, ok bool)

	// callbackParam returns the declarations of the exported function's
	// parameter anm, of type sym, in the cgo shim and in build.py, if it is
	// a callback declared unlike any other parameter (ok true).
	callbackParam(g *pyGen, anm string, sym *symbol) (goArg, pyArg string, ok bool)
	// callbackPrologue returns the Go statements that set up the func-typed
	// argument anm, at the start of the exported function.
	callbackPrologue(anm string) string
	// callbackArg returns the Go expression passed to the bound function
	// for the func-typed argument anm, of type sym.
	callbackArg(g *pyGen, anm string, sym *symbol) string

	// saveThread returns the Go statements that release the GIL ahead of
	// the call, and restore it when the exported function returns if
	// deferRestore is set; otherwise restoreThread's must run instead.
	saveThread(deferRestore bool) string
	// restoreThread returns the Go statements that reacquire the GIL.
	restoreThread() string
	// setError returns the Go statements that raise a Python exception exc
	// (the name of a builtin exception) with a message given both as the
	// Go string expression goMsg and the *C.char expression cMsg.
	setError(exc, goMsg, cMsg string) string
	// genBytesFuncs writes the functions that convert Slice_byte from and
	// to python bytes.
	genBytesFuncs(g *pyGen)
}

var (
	_ shimABI = cpythonShim{}
	_ shimABI = plainCShim{}
)

// newShim returns the shim that be generates.
func newShim(be backendImpl) shimABI {
	if nb, ok := be.(plainCBackend); ok {
		return plainCShim{cb: nb}
	}
	return cpythonShim{}
}

// cpythonShim is the shim of the pybindgen and capi backends, which calls
// the CPython C API itself.
type cpythonShim struct{}

func (cpythonShim) genGoPreamble(g *pyGen, pkgimport string) {
	libcfg := func() string {
		pycfg, err := GetPythonConfig(g.cfg.VM)
		if err != nil {
			panic(err)
		}
		var ldflags string
		if g.mode == ModeExe || !g.dynamicLink {
			ldflags = pycfg.LdFlags
		} else {
			ldflags = pycfg.LdDynamicFlags
		}
		// this is critical to avoid pybindgen errors:
		exflags := " -Wno-error -Wno-implicit-function-declaration -Wno-int-conversion"
		pkgcfg := fmt.Sprintf(`
#cgo CFLAGS: %s
#cgo LDFLAGS: %s
`, pycfg.CFlags+exflags, ldflags)

		return pkgcfg
	}()

	if g.mode == ModeExe && g.cfg.Main == "" {
		g.cfg.Main = "GoPyMainRun()" // default is just to run main
	}
	exeprec := ""
	exeprego := ""
	if g.mode == ModeExe {
		exeprec = fmt.Sprintf(goExePreambleC, g.cfg.Name)
		exeprego = goExePreambleGo
	}
	g.gofile.Printf(goPreamble, g.cfg.Name, g.cfg.Cmd, libcfg, GoHandle, CGoHandle,
		pkgimport, g.cfg.Main, exeprec, exeprego, g.cfg.Version)
}

func (cpythonShim) finishGo(g *pyGen) {}

func (cpythonShim) pyGenerate(g *pyGen) string {
	return fmt.Sprintf("\nmod.generate(open('%v.c', 'w'))\n\n", g.cfg.Name)
}

func (cpythonShim) funcSupported(g *pyGen, args, res []*Var) bool { return true }

func (cpythonShim) sliceSupported(esym *symbol) bool { return true }

func (cpythonShim) genComplexFunc(g *pyGen, fsym *Func) bool { return false }

func (cpythonShim) complexParts(sym *symbol) (string, bool) { return "", false }

func (cpythonShim) callbackParam(g *pyGen, anm string, sym *symbol) (string, string, bool) {
	return "", "", false
}

func (cpythonShim) callbackPrologue(anm string) string {
	return fmt.Sprintf("_fun_arg := %s\n", anm)
}

func (cpythonShim) callbackArg(g *pyGen, anm string, sym *symbol) string {
	return sym.py2go
}

func (cpythonShim) saveThread(deferRestore bool) string {
	s := "_saved_thread := C.PyEval_SaveThread()\n"
	if deferRestore {
		s += "defer C.PyEval_RestoreThread(_saved_thread)\n"
	}
	return s
}

func (cpythonShim) restoreThread() string {
	return "C.PyEval_RestoreThread(_saved_thread)\n"
}

func (cpythonShim) setError(exc, goMsg, cMsg string) string {
	return fmt.Sprintf("C.PyErr_SetString(C.PyExc_%s, %s)\n", exc, cMsg)
}

// goBytesFuncsCPython is cpythonShim's Slice_byte_from_bytes and
// Slice_byte_to_bytes; %s is the cgo type of a Py_ssize_t.
const goBytesFuncsCPython = `//export Slice_byte_from_bytes
func Slice_byte_from_bytes(o *C.PyObject) CGoHandle {
	size := C.PyBytes_Size(o)
	ptr := unsafe.Pointer(C.PyBytes_AsString(o))
	data := make([]byte, size)
	tmp := unsafe.Slice((*byte)(ptr), size)
	copy(data, tmp)
	return handleFromPtr_Slice_byte(&data)
}

//export Slice_byte_to_bytes
func Slice_byte_to_bytes(handle CGoHandle) *C.PyObject {
	s := deptrFromHandle_Slice_byte(handle)
	ptr := unsafe.Pointer(&s[0])
	size := len(s)
	return C.PyBytes_FromStringAndSize((*C.char)(ptr), %s(size))
}

`

func (cpythonShim) genBytesFuncs(g *pyGen) {
	ssize := "C.long"
	if WindowsOS {
		ssize = "C.longlong"
	}
	g.gofile.Printf(goBytesFuncsCPython, ssize)

	g.pybuild.Printf("mod.add_function('Slice_byte_from_bytes', retval('%s'), [param('PyObject*', 'o', transfer_ownership=False)])\n", PyHandle)
	g.pybuild.Printf("mod.add_function('Slice_byte_to_bytes', retval('PyObject*', caller_owns_return=True), [param('%s', 'handle')])\n", PyHandle)
}
