// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/parser"
	"go/types"
	"testing"
)

func TestCallbackTrampolines(t *testing.T) {
	cbs := []*callbackSig{
		{params: []callbackParam{{ctype: cbInt64}, {ctype: cbBool}}, ret: &callbackResult{ctype: cbDouble}},
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
			{name: "i", gotyp: "int", ctype: cbInt64, conv: "C.int64_t(i)"},
			{name: "s", gotyp: "string", ctype: cbCString, conv: "_cs", pre: "_cs := C.CString(string(s))\n"},
		},
		ret: &callbackResult{gotyp: "bool", ctype: cbBool},
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
	intArg := func() *callbackSig { return &callbackSig{params: []callbackParam{{ctype: cbInt64}}} }
	noArg := &callbackSig{}
	for i, tc := range []struct {
		cb   *callbackSig
		want int
	}{
		{intArg(), 0},
		{noArg, 1},
		{intArg(), 0}, // a distinct callbackSig of the same shape shares its trampoline
		{&callbackSig{ret: &callbackResult{ctype: cbInt64}}, 2},
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

func TestCbCType(t *testing.T) {
	// String is what each plain-C build.py parses, so it must not change.
	for _, tc := range []struct {
		t      cbCType
		str, c string
		kinds  []types.BasicKind
	}{
		{cbInt64, "int64_t", "int64_t", []types.BasicKind{types.Int, types.Int8, types.Int64}},
		{cbUint64, "uint64_t", "uint64_t", []types.BasicKind{types.Uint, types.Uint8, types.Uintptr}},
		{cbDouble, "double", "double", []types.BasicKind{types.Float32, types.Float64}},
		{cbBool, "bool", "uint8_t", []types.BasicKind{types.Bool}},
		{cbCString, "char*", "char*", []types.BasicKind{types.String}},
	} {
		if got := tc.t.String(); got != tc.str {
			t.Errorf("%d.String() = %q, want %q", uint8(tc.t), got, tc.str)
		}
		if got := tc.t.c(); got != tc.c {
			t.Errorf("%s.c() = %q, want %q", tc.t, got, tc.c)
		}
		for _, k := range tc.kinds {
			if got, ok := callbackBasicCType(k); !ok || got != tc.t {
				t.Errorf("callbackBasicCType(%v) = %s, %v; want %s", k, got, ok, tc.t)
			}
		}
	}
	for _, k := range []types.BasicKind{types.Complex64, types.Complex128, types.UnsafePointer} {
		if got, ok := callbackBasicCType(k); ok {
			t.Errorf("callbackBasicCType(%v) = %s, want none", k, got)
		}
	}
	if got := cbCType(0).String(); got != "cbCType(0)" {
		t.Errorf("zero cbCType.String() = %q", got)
	}
}
