// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import "fmt"

// A "no-API" backend generates exported Go functions that never touch the
// CPython C API, because their python-facing wrapper is produced outside the
// Go build: cffi loads a plain shared library at runtime; pybind11 compiles
// a C++ file against that same library.  This file holds what the two share
// about the shape of that library -- backend detection, error reporting
// (gopySetError, defined in cffi.go's preamble and used by both), and how a
// complex64/128 value crosses the boundary -- leaving each backend's own file
// for what only it needs (cffiBuildPreamble in cffi.go, pybind11BuildPreamble
// in pybind11.go).

func (g *pyGen) isCFFI() bool {
	return g.cfg.Backend == BackendCFFI
}

func (g *pyGen) isPyBind11() bool {
	return g.cfg.Backend == BackendPyBind11
}

// noAPIShim reports whether the exported Go functions must avoid the
// CPython C API (see the file comment above).
func (g *pyGen) noAPIShim() bool {
	return g.isCFFI() || g.isPyBind11()
}

// goSetError returns Go code that records an error for Python to raise.
// kind is the name of a Python builtin exception, msg a Go string expression.
func (g *pyGen) goSetError(kind, msg string) string {
	return "gopySetError(\"" + kind + "\", " + msg + ")\n"
}

func isComplexSym(sym *symbol) bool {
	return sym != nil && (sym.goname == "complex64" || sym.goname == "complex128")
}

// A complex64/complex128 value has no single C type that cffi or pybind11
// can declare (cgo's is _Complex), and cgo won't export a struct, so under
// either it crosses as two floats: as two parameters (<name>_re, <name>_im),
// and as a result in cgo's two-value return, which it exports as a plain C
// struct {r0; r1;}.  The methods below say how a value of a given symbol
// crosses, so that the generators only differ from the default backend here.

// isComplexShim reports whether sym crosses as two floats, under either
// no-API backend (cffi or pybind11).
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
// wrapper in cffi_build.py and pybind11_build.py expands complex64/128.
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
