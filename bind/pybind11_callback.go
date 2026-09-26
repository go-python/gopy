// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
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
// The parameter/result type rules are shared with cffi (cffiCallback,
// cffiCallbackParam, cffiCallbackResult in cffi_callback.go): the same Go
// types are supported, crossing as the same int64_t/uint64_t/double/bool/
// char* vocabulary either way.

// pybind11TrampolinesKey stands in for the extern declarations in the cgo
// preamble, which are written before the callback types that need them are
// known.  Unlike cffiTrampolinesKey, these are declarations only: the
// trampolines themselves are defined in the .cpp pybind11_build.py writes
// (callback_trampoline), not here -- see the buildPyBind11 doc comment
// (cmd_build.go) for why Go and that .cpp can't be two separate libraries
// with a dependency in each direction.
const pybind11TrampolinesKey = "@@GOPY_PYBIND11_TRAMPOLINES@@"

// splicePyBind11Trampolines writes the extern declarations into the cgo
// preamble, so the C compiler accepts calls to a function it never sees
// defined; the actual gopy_cb_N functions are resolved at the final link
// step in buildPyBind11, against the object code pybind11_build.py's
// generated .cpp compiles to.
func (g *pyGen) splicePyBind11Trampolines() {
	if !g.isPyBind11() {
		return
	}
	var c strings.Builder
	for i, cb := range g.cbs {
		params := []string{"int64_t h"}
		for j, p := range cb.params {
			params = append(params, fmt.Sprintf("%s a%d", cffiCType(p.ctype), j))
		}
		ret := "void"
		if cb.ret != nil {
			ret = cffiCType(cb.ret.ctype)
		}
		fmt.Fprintf(&c, "extern %s gopy_cb_%d(%s);\n", ret, i, strings.Join(params, ", "))
	}
	b := bytes.Replace(g.gofile.buf.Bytes(), []byte(pybind11TrampolinesKey), []byte(c.String()), 1)
	g.gofile.buf = bytes.NewBuffer(b)
}

// pybind11CallbackLit returns a Go func literal that calls the Python
// callable registered under the handle named anm.
func (g *pyGen) pybind11CallbackLit(cb *cffiCallback, anm string) string {
	var decl, pre []string
	args := []string{"C.int64_t(" + anm + ")"}
	for _, p := range cb.params {
		decl = append(decl, p.name+" "+p.gotyp)
		if p.pre != "" {
			pre = append(pre, p.pre)
		}
		args = append(args, p.conv)
	}
	call := fmt.Sprintf("C.gopy_cb_%d(%s)", g.cffiTrampoline(cb), strings.Join(args, ", "))
	result := ""
	if cb.ret != nil {
		result = " " + cb.ret.gotyp
		if cb.ret.ctype == "bool" {
			call += " != 0"
		}
		call = "return " + cb.ret.gotyp + "(" + call + ")"
	}
	return fmt.Sprintf("func(%s)%s {\n%s%s\n}", strings.Join(decl, ", "), result, strings.Join(pre, ""), call)
}
