// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

// A Python callable passed to Go as a func-typed argument crosses the cffi
// boundary as a C function pointer: the python side wraps the callable in an
// ffi.callback (see callback_setup in cffi_build.py), and the Go closure
// built here calls that pointer through a small static C trampoline, since
// cgo cannot call a C function pointer directly.
//
// The callback only lives as long as the Go call it was passed to, so the
// closure goes through a gopyCallbackScope that is closed when that call
// returns, instead of jumping to freed memory if Go kept it any longer.
//
// Which callback types are supported, and how their parameters and result
// cross, is shared with the C++ backends (see callback.go).

// cbGoParam returns unsafe.Pointer: the callable crosses as the C function
// pointer of its ffi.callback.
func (cffiBackend) cbGoParam() string { return "unsafe.Pointer" }

// cbPrologue keeps the function pointer as _cbfp_<anm>, and opens the
// gopyCallbackScope _cbs_<anm> that cbLit's closure enters, closing it when
// the exported function returns.
func (cffiBackend) cbPrologue(anm string) string {
	return fmt.Sprintf("_cbfp_%[1]s := %[1]s\n_cbs_%[1]s := new(gopyCallbackScope)\ndefer _cbs_%[1]s.close()\n", anm)
}

// cbLit returns a closure that calls the function pointer through
// trampoline n while _cbs_<anm> is open, and otherwise does nothing (and
// returns the zero value).
func (cffiBackend) cbLit(cb *callbackSig, n int, anm string) string {
	var decl, pre []string
	args := []string{"_cbfp_" + anm}
	for _, p := range cb.params {
		decl = append(decl, p.name+" "+p.gotyp)
		pre = append(pre, p.pre)
		args = append(args, p.conv)
	}
	call := fmt.Sprintf("C.gopy_cb_%d(%s)", n, strings.Join(args, ", "))
	result := ""
	if cb.ret != nil {
		// named, so that a refused call returns its zero value
		result = " (_r " + cb.ret.gotyp + ")"
		if cb.ret.ctype == "bool" {
			call += " != 0"
		}
		call = "return " + cb.ret.gotyp + "(" + call + ")"
	}
	return fmt.Sprintf("func(%[1]s)%[2]s {\nif !_cbs_%[3]s.enter() {\nreturn\n}\ndefer _cbs_%[3]s.leave()\n%[4]s%[5]s\n}",
		strings.Join(decl, ", "), result, anm, strings.Join(pre, ""), call)
}

// cbTrampolines defines a static C function per callback type, that calls
// the function pointer it is given.
func (cffiBackend) cbTrampolines(cbs []*callbackSig) string {
	var c strings.Builder
	for i, cb := range cbs {
		params := []string{"void* f"}
		ptypes := []string{}
		args := []string{}
		for j, p := range cb.params {
			t := callbackCType(p.ctype)
			params = append(params, fmt.Sprintf("%s a%d", t, j))
			ptypes = append(ptypes, t)
			args = append(args, fmt.Sprintf("a%d", j))
		}
		if len(ptypes) == 0 {
			ptypes = append(ptypes, "void")
		}
		ret, retStmt := "void", ""
		if cb.ret != nil {
			ret, retStmt = callbackCType(cb.ret.ctype), "return "
		}
		fmt.Fprintf(&c, "static inline %s gopy_cb_%d(%s) { %s((%s (*)(%s))f)(%s); }\n",
			ret, i, strings.Join(params, ", "), retStmt, ret, strings.Join(ptypes, ", "), strings.Join(args, ", "))
	}
	return c.String()
}
