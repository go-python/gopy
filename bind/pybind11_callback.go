// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

// A Python callable passed to Go as a func-typed argument crosses the
// pybind11 boundary as an int64 handle, not a raw C function pointer: unlike
// cffi (whose ffi.callback creates one real function pointer per Python
// callable, via a libffi closure allocated at runtime), pybind11 has no way
// to synthesize new C-ABI function pointers at runtime, so instead one
// static C++ trampoline exists per callback signature, shared by every
// callable of that shape, and the actual py::function is looked up from a
// registry by handle each time Go calls back in (see MODULE_TEMPLATE and
// callback_trampoline in pybind11_build.py).  Registering/unregistering the
// handle around the call (there, not here) is what makes a callback stop
// working once the python call it was passed to returns, playing the same
// role gopyCallbackScope plays for cffi.
//
// The parameter/result type rules are shared with cffi (callbackSig in
// callback.go): the same Go types are supported, crossing as the same
// int64_t/uint64_t/double/bool/char* vocabulary either way.
//
// Everything in this file serves the nanobind backend too (see cxxBackend),
// whose nanobind_build.py defines the same registry and trampolines.

// cbGoParam returns CGoHandle: the callable crosses as its registry handle.
func (cxxBackend) cbGoParam() string { return "CGoHandle" }

// cbPrologue returns nothing: there is no Go-side setup, since the C++
// registry itself refuses a call once the wrapping python call unregisters
// the handle.
func (cxxBackend) cbPrologue(anm string) string { return "" }

// cbLit returns a closure that calls trampoline n with the handle, which
// looks up the callable in the registry.
func (cxxBackend) cbLit(cb *callbackSig, n int, anm string) string {
	var decl, pre []string
	args := []string{"C.int64_t(" + anm + ")"}
	for _, p := range cb.params {
		decl = append(decl, p.name+" "+p.gotyp)
		if p.pre != "" {
			pre = append(pre, p.pre)
		}
		args = append(args, p.conv)
	}
	call := fmt.Sprintf("C.gopy_cb_%d(%s)", n, strings.Join(args, ", "))
	result := ""
	if cb.ret != nil {
		result = " " + cb.ret.gotyp
		if cb.ret.ctype == cbBool {
			call += " != 0"
		}
		call = "return " + cb.ret.gotyp + "(" + call + ")"
	}
	return fmt.Sprintf("func(%s)%s {\n%s%s\n}", strings.Join(decl, ", "), result, strings.Join(pre, ""), call)
}

// cbTrampolines writes extern declarations only, so the C compiler accepts
// calls to a function it never sees defined: the trampolines themselves are
// defined in the .cpp build.py writes (callback_trampoline in
// pybind11_build.py), and resolved at the final link step in
// buildCXXModule -- see the buildCXXModule doc comment (cmd_build.go) for
// why Go and that .cpp can't be two separate libraries with a dependency in
// each direction.
func (cxxBackend) cbTrampolines(cbs []*callbackSig) string {
	var c strings.Builder
	for i, cb := range cbs {
		params := []string{"int64_t h"}
		for j, p := range cb.params {
			params = append(params, fmt.Sprintf("%s a%d", p.ctype.c(), j))
		}
		ret := "void"
		if cb.ret != nil {
			ret = cb.ret.ctype.c()
		}
		fmt.Fprintf(&c, "extern %s gopy_cb_%d(%s);\n", ret, i, strings.Join(params, ", "))
	}
	return c.String()
}
