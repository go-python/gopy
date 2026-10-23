// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/parser"
	"testing"
)

func TestCallbackTrampolines(t *testing.T) {
	cbs := []*callbackSig{
		{params: []callbackParam{{ctype: "int64_t"}, {ctype: "bool"}}, ret: &callbackResult{ctype: "double"}},
		{},
	}
	for _, tc := range []struct {
		abi  callbackABI
		want string
	}{
		{cffiBackend{}, "static inline double gopy_cb_0(void* f, int64_t a0, uint8_t a1) { return ((double (*)(int64_t, uint8_t))f)(a0, a1); }\n" +
			"static inline void gopy_cb_1(void* f) { ((void (*)(void))f)(); }\n"},
		{pybind11Backend, "extern double gopy_cb_0(int64_t h, int64_t a0, uint8_t a1);\n" +
			"extern void gopy_cb_1(int64_t h);\n"},
	} {
		if got := tc.abi.cbTrampolines(cbs); got != tc.want {
			t.Errorf("%T.cbTrampolines:\ngot:\n%s\nwant:\n%s", tc.abi, got, tc.want)
		}
	}
}

func TestCallbackLit(t *testing.T) {
	full := &callbackSig{
		params: []callbackParam{
			{name: "i", gotyp: "int", ctype: "int64_t", conv: "C.int64_t(i)"},
			{name: "s", gotyp: "string", ctype: "char*", conv: "_cs", pre: "_cs := C.CString(string(s))\n"},
		},
		ret: &callbackResult{gotyp: "bool", ctype: "bool"},
	}
	empty := &callbackSig{}
	for _, tc := range []struct {
		abi  callbackABI
		cb   *callbackSig
		n    int
		anm  string
		want string
	}{
		{cffiBackend{}, full, 3, "fun", "func(i int, s string) (_r bool) {\n" +
			"if !_cbs_fun.enter() {\nreturn\n}\ndefer _cbs_fun.leave()\n" +
			"_cs := C.CString(string(s))\n" +
			"return bool(C.gopy_cb_3(_cbfp_fun, C.int64_t(i), _cs) != 0)\n}"},
		{cffiBackend{}, empty, 0, "f", "func() {\n" +
			"if !_cbs_f.enter() {\nreturn\n}\ndefer _cbs_f.leave()\n" +
			"C.gopy_cb_0(_cbfp_f)\n}"},
		{pybind11Backend, full, 3, "fun", "func(i int, s string) bool {\n" +
			"_cs := C.CString(string(s))\n" +
			"return bool(C.gopy_cb_3(C.int64_t(fun), C.int64_t(i), _cs) != 0)\n}"},
		{pybind11Backend, empty, 0, "f", "func() {\nC.gopy_cb_0(C.int64_t(f))\n}"},
	} {
		got := tc.abi.cbLit(tc.cb, tc.n, tc.anm)
		if got != tc.want {
			t.Errorf("%T.cbLit(%q):\ngot:\n%s\nwant:\n%s", tc.abi, tc.anm, got, tc.want)
		}
		if _, err := parser.ParseExpr(got); err != nil {
			t.Errorf("%T.cbLit(%q) is not a Go expression: %v\n%s", tc.abi, tc.anm, err, got)
		}
	}
}

func TestTrampolineNumbering(t *testing.T) {
	g := &pyGen{}
	intArg := func() *callbackSig { return &callbackSig{params: []callbackParam{{ctype: "int64_t"}}} }
	noArg := &callbackSig{}
	for i, tc := range []struct {
		cb   *callbackSig
		want int
	}{
		{intArg(), 0},
		{noArg, 1},
		{intArg(), 0}, // a distinct callbackSig of the same shape shares its trampoline
		{&callbackSig{ret: &callbackResult{ctype: "int64_t"}}, 2},
		{noArg, 1},
	} {
		if got := g.trampoline(tc.cb); got != tc.want {
			t.Errorf("call %d: trampoline(%s) = %d, want %d", i, tc.cb.pyType(), got, tc.want)
		}
	}
	if len(g.cbs) != 3 {
		t.Errorf("len(g.cbs) = %d, want 3", len(g.cbs))
	}
}
