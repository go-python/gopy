// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"fmt"
)

// A "plain-C" backend's cgo shim exports plain C functions, which never touch
// the CPython C API, because their python-facing wrapper is produced outside
// the Go build: cffi loads a plain shared library at runtime; pybind11 and
// nanobind compile a C++ file against that same library.  This file holds
// what they share about the shape of that library -- its cgo preamble
// (goPreamblePlainC, below), error reporting (gopySetError), and how a
// complex64/128 value crosses the boundary -- as plainCShim, leaving each
// backend's own file for what only it needs (cffi.go, pybind11.go,
// nanobind.go).

// callbackABI is how a python callable passed to Go as a func-typed argument
// crosses a plain-C shim: as a C function pointer for cffi (see
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

// plainCShim is the shim of the cffi, pybind11 and nanobind backends, which
// avoids the CPython C API (see the file comment above).  cb is how the
// backend passes callbacks.
type plainCShim struct {
	cb callbackABI
}

func (plainCShim) genGoPreamble(g *pyGen, pkgimport string) {
	g.gofile.Printf(goPreamblePlainC, g.cfg.Name, g.cfg.Cmd, GoHandle, CGoHandle,
		pkgimport, g.cfg.Main, trampolinesKey, g.cfg.Version)
}

// finishGo writes the callback trampolines in place of trampolinesKey.
func (s plainCShim) finishGo(g *pyGen) {
	b := bytes.Replace(g.gofile.buf.Bytes(), []byte(trampolinesKey), []byte(s.cb.cbTrampolines(g.cbs)), 1)
	g.gofile.buf = bytes.NewBuffer(b)
}

func (plainCShim) pyGenerate(g *pyGen) string {
	return "\nmod.generate()\n\n"
}

// funcSupported rejects a raw PyObject* (complex64/128, which
// genComplexFunc handles in a narrow case, and a callback argument of a type
// callbackSig doesn't handle), rather than emit a signature referencing the
// CPython C API, which would fail to even compile under goPreamblePlainC.
func (plainCShim) funcSupported(g *pyGen, args, res []*Var) bool {
	for _, arg := range args {
		sarg := current.symtype(arg.GoType())
		switch {
		case sarg == nil:
		case sarg.isSignature():
			if g.callbackSig(sarg) == nil {
				return false
			}
		case sarg.cpyname == "PyObject*":
			return false
		}
	}
	for _, ret := range res {
		if sret := current.symtype(ret.GoType()); sret != nil && sret.cpyname == "PyObject*" {
			return false
		}
	}
	return true
}

// sliceSupported rejects an element type that crosses as a raw PyObject*,
// for the same reason as funcSupported; element access (elem/set/append)
// would reference *C.PyObject, which goPreamblePlainC doesn't declare.
func (plainCShim) sliceSupported(esym *symbol) bool {
	return esym == nil || esym.cpyname != "PyObject*" || isComplexSym(esym)
}

func (plainCShim) genComplexFunc(g *pyGen, fsym *Func) bool {
	return g.genFuncComplexPlainC(fsym)
}

// A complex64/complex128 value has no single C type that cffi or a C++
// module can declare (cgo's is _Complex), and cgo won't export a struct, so
// under any of them it crosses as two floats: as two parameters (<name>_re, <name>_im),
// and as a result in cgo's two-value return, which it exports as a plain C
// struct {r0; r1;}.
func (plainCShim) complexParts(sym *symbol) (string, bool) {
	switch {
	case sym == nil:
		return "", false
	case sym.goname == "complex64":
		return "C.float", true
	case sym.goname == "complex128":
		return "C.double", true
	}
	return "", false
}

func (s plainCShim) callbackParam(g *pyGen, anm string, sym *symbol) (string, string, bool) {
	if !sym.isSignature() {
		return "", "", false
	}
	goArg := fmt.Sprintf("%s %s", anm, s.cb.cbGoParam())
	pyArg := fmt.Sprintf("param('%s', '%s')", g.callbackSig(sym).pyType(), anm)
	return goArg, pyArg, true
}

func (s plainCShim) callbackPrologue(anm string) string {
	return s.cb.cbPrologue(anm)
}

func (s plainCShim) callbackArg(g *pyGen, anm string, sym *symbol) string {
	cb := g.callbackSig(sym)
	return s.cb.cbLit(cb, g.trampoline(cb), anm)
}

// saveThread and restoreThread return nothing: cffi, pybind11 and nanobind
// each release the GIL themselves around every call.
func (plainCShim) saveThread(deferRestore bool) string { return "" }

func (plainCShim) restoreThread() string { return "" }

// setError records the error for the python side to raise (see
// gopySetError in goPreamblePlainC).
func (plainCShim) setError(exc, goMsg, cMsg string) string {
	return "gopySetError(\"" + exc + "\", " + goMsg + ")\n"
}

// genBytesFuncs exchanges a raw pointer+length instead of a PyObject*, since
// PyBytes_* is off-limits (no CPython headers); each plain-C build.py binds
// them by name (BYTES_FUNCS in cffi_build.py, MODULE_TEMPLATE in
// pybind11_build.py and nanobind_build.py) to convert from and to python
// bytes.
func (plainCShim) genBytesFuncs(g *pyGen) {
	g.gofile.Printf("//export Slice_byte_from_bytes\n")
	g.gofile.Printf("func Slice_byte_from_bytes(ptr unsafe.Pointer, size C.longlong) CGoHandle {\n")
	g.gofile.Indent()
	g.gofile.Printf("data := make([]byte, size)\n")
	g.gofile.Printf("if size > 0 {\n")
	g.gofile.Indent()
	g.gofile.Printf("tmp := unsafe.Slice((*byte)(ptr), size)\n")
	g.gofile.Printf("copy(data, tmp)\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("return handleFromPtr_Slice_byte(&data)\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.gofile.Printf("//export Slice_byte_to_bytes_len\n")
	g.gofile.Printf("func Slice_byte_to_bytes_len(handle CGoHandle) C.longlong {\n")
	g.gofile.Indent()
	g.gofile.Printf("s := deptrFromHandle_Slice_byte(handle)\n")
	g.gofile.Printf("return C.longlong(len(s))\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	// Returning &s[0] directly would hand cgo a pointer into the Go
	// heap, which cgo's pointer checks reject once it crosses back
	// to the caller; copy into a C-owned buffer instead, freed by
	// the python side (Slice_byte_free_ptr) once it has read it.
	g.gofile.Printf("//export Slice_byte_to_bytes_ptr\n")
	g.gofile.Printf("func Slice_byte_to_bytes_ptr(handle CGoHandle) unsafe.Pointer {\n")
	g.gofile.Indent()
	g.gofile.Printf("s := deptrFromHandle_Slice_byte(handle)\n")
	g.gofile.Printf("n := len(s)\n")
	g.gofile.Printf("if n == 0 {\n")
	g.gofile.Indent()
	g.gofile.Printf("return nil\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n")
	g.gofile.Printf("buf := C.malloc(C.size_t(n))\n")
	g.gofile.Printf("copy(unsafe.Slice((*byte)(buf), n), s)\n")
	g.gofile.Printf("return buf\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.gofile.Printf("//export Slice_byte_free_ptr\n")
	g.gofile.Printf("func Slice_byte_free_ptr(ptr unsafe.Pointer) {\n")
	g.gofile.Indent()
	g.gofile.Printf("C.free(ptr)\n")
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")
}

func isComplexSym(sym *symbol) bool {
	return sym != nil && (sym.goname == "complex64" || sym.goname == "complex128")
}

// The methods below say how a value of a given symbol crosses the shim, so
// that the generators only differ by backend where shimABI.complexParts says
// it crosses as two floats.

// cgoParam returns the declaration of the parameter of an exported function
// that carries a value of sym.
func (g *pyGen) cgoParam(name string, sym *symbol) string {
	if cf, ok := g.shim.complexParts(sym); ok {
		return fmt.Sprintf("%[1]s_re %[2]s, %[1]s_im %[2]s", name, cf)
	}
	return name + " " + sym.cgoname
}

// cgoResult returns the result type of an exported function that returns a
// value of sym.
func (g *pyGen) cgoResult(sym *symbol) string {
	if cf, ok := g.shim.complexParts(sym); ok {
		return "(" + cf + ", " + cf + ")"
	}
	return sym.cgoname
}

// cpyName returns the type that build.py records for a value of sym.
// wrapper in cffi_build.py, pybind11_build.py and nanobind_build.py expands
// complex64/128.
func (g *pyGen) cpyName(sym *symbol) string {
	if _, ok := g.shim.complexParts(sym); ok {
		return sym.goname
	}
	return sym.cpyname
}

// goToCgo returns the Go expression that converts expr, a value of sym, to
// what an exported function returns.
func (g *pyGen) goToCgo(sym *symbol, expr string) string {
	if _, ok := g.shim.complexParts(sym); ok {
		return sym.goname + "GoToPyPlainC(" + expr + ")"
	}
	if sym.go2py != "" {
		return sym.go2py + "(" + expr + ")" + sym.go2pyParenEx
	}
	return expr
}

// cgoToGo returns the Go expression that converts the parameter name, as
// declared by cgoParam, to a value of sym.
func (g *pyGen) cgoToGo(sym *symbol, name string) string {
	if _, ok := g.shim.complexParts(sym); ok {
		return sym.goname + "PyToGoPlainC(" + name + "_re, " + name + "_im)"
	}
	if sym.py2go != "" {
		return sym.py2go + "(" + name + ")" + sym.py2goParenEx
	}
	return name
}

// goPreamblePlainC is the cgo preamble of every plain-C backend's shim:
// 1 = name of package, 2 = cmdstr, 3 = GoHandle, 4 = CGoHandle, 5 = all
// imports, 6 = mainstr, 7 = C trampolines for callbacks (see callbackABI),
// 8 = gopy version.
const goPreamblePlainC = `/*
cgo stubs for package %[1]s, for use with cffi, pybind11 or nanobind.
File is generated by gopy version %[8]s. Do not edit.
%[2]s
*/

package main

/*
#include <stdint.h>
#include <stdlib.h>
// pybind11's consumer #includes this same preamble (via the generated
// header) as C++, where bool/thread_local are keywords, not the C ones below.
#if !defined(__cplusplus) && (!defined(__STDC_VERSION__) || (__STDC_VERSION__ < 202311L))
typedef uint8_t bool;
#endif
#if defined(__cplusplus)
#define GOPY_THREAD_LOCAL thread_local
#else
#define GOPY_THREAD_LOCAL _Thread_local
#endif
// Thread-local storage for the last error, read by the python side right
// after the call that set it, on the same OS thread.  Windows gets its own
// branch: mingw's compiler-provided thread-local storage (_Thread_local)
// needs TLS support that a DLL loaded after process start (as cffi and
// ctypes do) isn't guaranteed to have, so this uses the OS API instead,
// which carries no such restriction (see TlsAlloc/TlsGetValue/TlsSetValue).
#ifdef _WIN32
#include <windows.h>
static DWORD gopy_err_tls = TLS_OUT_OF_INDEXES;
static inline void gopy_err_tls_init(void) { gopy_err_tls = TlsAlloc(); }
static inline char* gopy_err_get(void) { return (char*)TlsGetValue(gopy_err_tls); }
static inline void gopy_err_put(char* msg) { TlsSetValue(gopy_err_tls, msg); }
#else
static inline void gopy_err_tls_init(void) {}
static GOPY_THREAD_LOCAL char* gopy_err_msg;
static inline char* gopy_err_get(void) { return gopy_err_msg; }
static inline void gopy_err_put(char* msg) { gopy_err_msg = msg; }
#endif
static inline void gopy_set_err(char* msg) {
	free(gopy_err_get());
	gopy_err_put(msg);
}
static inline char* gopy_take_err(void) {
	char* msg = gopy_err_get();
	gopy_err_put(NULL);
	return msg;
}
%[7]s
*/
import "C"
import (
	"runtime"
	"sync"
	"unsafe"
	"github.com/go-python/gopy/gopyh" // handler
	%[5]s
)

func main() {
	%[6]s
}

//export GoPyInit
func GoPyInit() {
	%[6]s
}

// type for the handle -- int64 for speed (can switch to string)
type GoHandle %[3]s
type CGoHandle %[4]s

// DecRef decrements the reference count for the specified handle
// and deletes it it goes to zero.
//export DecRef
func DecRef(handle CGoHandle) {
	gopyh.DecRef(gopyh.CGoHandle(handle))
}

// IncRef increments the reference count for the specified handle.
//export IncRef
func IncRef(handle CGoHandle) {
	gopyh.IncRef(gopyh.CGoHandle(handle))
}

// NumHandles returns the number of handles currently in use.
//export NumHandles
func NumHandles() int {
	return gopyh.NumHandles()
}

// RequestGC runs Go's garbage collector on a dedicated goroutine, and waits.
//export RequestGC
func RequestGC() {
	done := make(chan struct{})
	_gcReq <- done
	<-done
}

var _gcReq = make(chan chan struct{})

func init() {
	C.gopy_err_tls_init()
	go func() {
		for done := range _gcReq {
			runtime.GC()
			close(done)
		}
	}()
}

// The error of the last call on the calling OS thread, for the python side
// to raise; gopy_err_msg is thread-local so concurrent calls on other
// threads don't see it.
func gopySetError(kind, msg string) {
	C.gopy_set_err(C.CString(kind + ":" + msg))
}

// GopyTakeError returns "Kind:message" for the error recorded by the last
// call on the calling thread and clears it, or NULL if there is none.
// Free with GopyFreeString.
//export GopyTakeError
func GopyTakeError() *C.char {
	return C.gopy_take_err()
}

// GopyFreeString frees a string returned by this library.
//export GopyFreeString
func GopyFreeString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

// gopyCallbackScope guards a python callback passed to Go, which only exists
// while the python call it was passed to is running (see cffi_callback.go):
// close waits for callbacks that are running, and then refuses new ones.
type gopyCallbackScope struct {
	mu   sync.RWMutex
	done bool
}

func (s *gopyCallbackScope) enter() bool {
	s.mu.RLock()
	if s.done {
		s.mu.RUnlock()
		println("gopy: callback called after the python call it was passed to returned")
		return false
	}
	return true
}

func (s *gopyCallbackScope) leave() { s.mu.RUnlock() }

func (s *gopyCallbackScope) close() {
	s.mu.Lock()
	s.done = true
	s.mu.Unlock()
}

// boolGoToPy converts a Go bool to python-compatible C.char
func boolGoToPy(b bool) C.char {
	if b {
		return 1
	}
	return 0
}

// boolPyToGo converts a python-compatible C.Char to Go bool
func boolPyToGo(b C.char) bool {
	return b != 0
}

// errorGoToPy converts a Go error to python-compatible C.CString
func errorGoToPy(e error) *C.char {
	if e != nil {
		return C.CString(e.Error())
	}
	return C.CString("")
}

// complex values cross as two floats, see plainCShim.complexParts in plainc.go
func complex64GoToPyPlainC(c complex64) (C.float, C.float) {
	return C.float(real(c)), C.float(imag(c))
}

func complex64PyToGoPlainC(re, im C.float) complex64 {
	return complex(float32(re), float32(im))
}

func complex128GoToPyPlainC(c complex128) (C.double, C.double) {
	return C.double(real(c)), C.double(imag(c))
}

func complex128PyToGoPlainC(re, im C.double) complex128 {
	return complex(float64(re), float64(im))
}
`
