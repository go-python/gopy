// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"fmt"
)

// A "no-API" backend generates exported Go functions that never touch the
// CPython C API, because their python-facing wrapper is produced outside the
// Go build: cffi loads a plain shared library at runtime; pybind11 and
// nanobind compile a C++ file against that same library.  This file holds
// what they share about the shape of that library -- backend detection,
// error reporting (gopySetError, defined in cffi.go's preamble and used by
// all of them), and how a complex64/128 value crosses the boundary --
// leaving each backend's own file for what only it needs (cffi.go,
// pybind11.go, nanobind.go).

// callbackABI is how a python callable passed to Go as a func-typed argument
// crosses a no-API shim: as a C function pointer for cffi (see
// cffi_callback.go), or as a registry handle for the C++ backends (see
// pybind11_callback.go).  Both support the same callback types (see
// callback.go).
type callbackABI interface {
	// cbGoParam returns the Go type of the exported function's parameter
	// that carries the callable.
	cbGoParam() string
	// cbPrologue returns the Go statements that set up the callable passed
	// as the argument named anm, ahead of the call cbLit's result is
	// passed to.
	cbPrologue(anm string) string
	// cbLit returns a Go func literal, of the callback type cb describes,
	// that calls the callable passed as the argument named anm through
	// trampoline n (gopy_cb_<n>, see pyGen.trampoline).
	cbLit(cb *callbackSig, n int, anm string) string
	// cbTrampolines returns the C code for the cgo preamble that lets Go
	// call each of cbs, numbered gopy_cb_<index>.
	cbTrampolines(cbs []*callbackSig) string
}

// trampolinesKey stands in for cbTrampolines' result in the cgo preamble,
// which is written before the callback types that need them are known.
const trampolinesKey = "@@GOPY_TRAMPOLINES@@"

// noAPI returns the backend if its shim must avoid the CPython C API (see
// the file comment above), or nil if it doesn't.
func (g *pyGen) noAPI() noAPIBackend {
	nb, _ := g.be.(noAPIBackend)
	return nb
}

// noAPIShim reports whether the exported Go functions must avoid the
// CPython C API (see the file comment above).
func (g *pyGen) noAPIShim() bool {
	return g.noAPI() != nil
}

// spliceTrampolines writes the callback trampolines in place of
// trampolinesKey, for a no-API backend.
func (g *pyGen) spliceTrampolines() {
	nb := g.noAPI()
	if nb == nil {
		return
	}
	b := bytes.Replace(g.gofile.buf.Bytes(), []byte(trampolinesKey), []byte(nb.cbTrampolines(g.cbs)), 1)
	g.gofile.buf = bytes.NewBuffer(b)
}

// goSetError returns Go code that records an error for Python to raise.
// kind is the name of a Python builtin exception, msg a Go string expression.
func (g *pyGen) goSetError(kind, msg string) string {
	return "gopySetError(\"" + kind + "\", " + msg + ")\n"
}

func isComplexSym(sym *symbol) bool {
	return sym != nil && (sym.goname == "complex64" || sym.goname == "complex128")
}

// A complex64/complex128 value has no single C type that cffi or a C++
// module can declare (cgo's is _Complex), and cgo won't export a struct, so
// under any of them it crosses as two floats: as two parameters (<name>_re, <name>_im),
// and as a result in cgo's two-value return, which it exports as a plain C
// struct {r0; r1;}.  The methods below say how a value of a given symbol
// crosses, so that the generators only differ from the default backend here.

// isComplexShim reports whether sym crosses as two floats, under either
// no-API backend (cffi, pybind11 or nanobind).
func (g *pyGen) isComplexShim(sym *symbol) bool {
	return g.noAPIShim() && isComplexSym(sym)
}

// cffiComplexFloat returns the cgo and the Go float type of the parts of a
// complex64 or complex128 symbol.
func cffiComplexFloat(sym *symbol) (cfloat, gofloat string) {
	if sym.goname == "complex64" {
		return "C.float", "float32"
	}
	return "C.double", "float64"
}

// cgoParam returns the declaration of the parameter of an exported function
// that carries a value of sym.
func (g *pyGen) cgoParam(name string, sym *symbol) string {
	if g.isComplexShim(sym) {
		cf, _ := cffiComplexFloat(sym)
		return fmt.Sprintf("%[1]s_re %[2]s, %[1]s_im %[2]s", name, cf)
	}
	return name + " " + sym.cgoname
}

// cgoResult returns the result type of an exported function that returns a
// value of sym.
func (g *pyGen) cgoResult(sym *symbol) string {
	if g.isComplexShim(sym) {
		cf, _ := cffiComplexFloat(sym)
		return "(" + cf + ", " + cf + ")"
	}
	return sym.cgoname
}

// cpyName returns the type that build.py records for a value of sym.
// wrapper in cffi_build.py, pybind11_build.py and nanobind_build.py expands
// complex64/128.
func (g *pyGen) cpyName(sym *symbol) string {
	if g.isComplexShim(sym) {
		return sym.goname
	}
	return sym.cpyname
}

// goToCgo returns the Go expression that converts expr, a value of sym, to
// what an exported function returns.
func (g *pyGen) goToCgo(sym *symbol, expr string) string {
	switch {
	case g.isComplexShim(sym):
		return sym.goname + "GoToPyCFFI(" + expr + ")"
	case sym.go2py != "":
		return sym.go2py + "(" + expr + ")" + sym.go2pyParenEx
	}
	return expr
}

// cgoToGo returns the Go expression that converts the parameter name, as
// declared by cgoParam, to a value of sym.
func (g *pyGen) cgoToGo(sym *symbol, name string) string {
	switch {
	case g.isComplexShim(sym):
		return sym.goname + "PyToGoCFFI(" + name + "_re, " + name + "_im)"
	case sym.py2go != "":
		return sym.py2go + "(" + name + ")" + sym.py2goParenEx
	}
	return name
}
