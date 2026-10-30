// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

// What a plain-C backend's callbacks share, whichever way the callable
// itself crosses (see callbackABI): which callback types are supported, and
// how each parameter and result crosses.
//
// Supported: parameters that are pointer/interface handles, numbers, bool,
// string or interface{} (which arrives as a string, as with pybindgen), and a
// result that is a number or bool.  Any other callback type makes the
// function be skipped, see genFuncSig.

// cbCType is the C type that a callback's parameter or result crosses as.
// The zero value is none of them.
type cbCType uint8

const (
	cbInt64   cbCType = iota + 1 // int64_t
	cbUint64                     // uint64_t
	cbDouble                     // double
	cbBool                       // bool, a byte in C code
	cbCString                    // char*, a parameter only
)

// String returns t's name in pyType's spec of a callback type, which each
// plain-C backend's build.py takes apart.
func (t cbCType) String() string {
	switch t {
	case cbInt64:
		return "int64_t"
	case cbUint64:
		return "uint64_t"
	case cbDouble:
		return "double"
	case cbBool:
		return "bool"
	case cbCString:
		return "char*"
	}
	return fmt.Sprintf("cbCType(%d)", uint8(t))
}

// c returns t's type in C code (and, as C.<c>, in cgo): bool crosses as a
// byte, since the cgo preamble can't rely on C's own bool.
func (t cbCType) c() string {
	if t == cbBool {
		return "uint8_t"
	}
	return t.String()
}

// callbackParam is one parameter of a callback.
type callbackParam struct {
	name  string  // name of the parameter in the func literal
	gotyp string  // its Go type
	ctype cbCType // how it crosses
	pre   string  // Go statements to run before the call, if any
	conv  string  // Go expression converting it to ctype
}

// callbackResult is the result of a callback.
type callbackResult struct {
	gotyp string  // its Go type
	ctype cbCType // how it crosses: never cbCString
}

// callbackSig is how the parameters and result of a callback type cross.
type callbackSig struct {
	params []callbackParam
	ret    *callbackResult // nil if the callback has no result
}

// callbackBasicCType returns how a value of a basic Go type crosses, and
// whether it can.
func callbackBasicCType(k types.BasicKind) (cbCType, bool) {
	switch {
	case types.Int <= k && k <= types.Int64:
		return cbInt64, true
	case types.Uint <= k && k <= types.Uintptr:
		return cbUint64, true
	case k == types.Float32 || k == types.Float64:
		return cbDouble, true
	case k == types.Bool:
		return cbBool, true
	case k == types.String:
		return cbCString, true
	}
	return 0, false
}

// callbackSig returns how to pass sym, a func-typed argument, to Go
// as a callback, or nil if its type isn't supported.
func (g *pyGen) callbackSig(sym *symbol) *callbackSig {
	sig, ok := sym.GoType().Underlying().(*types.Signature)
	if !ok || sig.Results().Len() > 1 || sig.Variadic() {
		return nil
	}
	cb := &callbackSig{}
	for i := 0; i < sig.Params().Len(); i++ {
		p, ok := newCallbackParam(sig.Params().At(i), i)
		if !ok {
			return nil
		}
		cb.params = append(cb.params, p)
	}
	if sig.Results().Len() == 1 {
		r, ok := newCallbackResult(sig.Results().At(0).Type())
		if !ok {
			return nil
		}
		cb.ret = r
	}
	return cb
}

// cStringPre returns the Go statements that make the C string _c<nm> from
// expr, for the duration of the call.  The python side copies it.
func cStringPre(nm, expr string) string {
	return fmt.Sprintf("_c%[1]s := %[2]s\ndefer C.free(unsafe.Pointer(_c%[1]s))\n", nm, expr)
}

func newCallbackParam(v *types.Var, i int) (callbackParam, bool) {
	typ := v.Type()
	vsym := current.symtype(typ)
	if vsym == nil {
		return callbackParam{}, false
	}
	nm := pySafeArg(v.Name(), i)
	p := callbackParam{name: nm, gotyp: current.typeGoName(typ)}

	if vsym.hasHandle() && vsym.isPtrOrIface() {
		p.ctype = cbInt64
		p.conv = fmt.Sprintf("C.int64_t(%s(%s)%s)", vsym.go2py, nm, vsym.go2pyParenEx)
		return p, true
	}
	if vsym.goname == "interface{}" {
		p.ctype, p.conv = cbCString, "_c"+nm
		p.pre = cStringPre(nm, fmt.Sprintf("%s(%s)%s", vsym.go2py, nm, vsym.go2pyParenEx))
		return p, true
	}
	bt, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return p, false
	}
	if p.ctype, ok = callbackBasicCType(bt.Kind()); !ok {
		return p, false
	}
	switch p.ctype {
	case cbBool:
		p.conv = fmt.Sprintf("C.%s(boolGoToPy(bool(%s)))", p.ctype.c(), nm)
	case cbCString:
		p.conv = "_c" + nm
		p.pre = cStringPre(nm, fmt.Sprintf("C.CString(string(%s))", nm))
	default:
		p.conv = fmt.Sprintf("C.%s(%s)", p.ctype.c(), nm)
	}
	return p, true
}

func newCallbackResult(typ types.Type) (*callbackResult, bool) {
	bt, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return nil, false
	}
	ctype, ok := callbackBasicCType(bt.Kind())
	if !ok || ctype == cbCString { // the python side has no way to give Go a string it owns
		return nil, false
	}
	return &callbackResult{gotyp: current.typeGoName(typ), ctype: ctype}, true
}

// pyType is the type given to the callback's parameter in build.py, which
// each plain-C backend's build.py takes apart (callback_setup in
// cffi_build.py, callback_trampoline in pybind11_build.py and
// nanobind_build.py):
// callback:<result or void>(<parameter types>)
func (cb *callbackSig) pyType() string {
	ret := "void"
	if cb.ret != nil {
		ret = cb.ret.ctype.String()
	}
	ts := make([]string, len(cb.params))
	for i, p := range cb.params {
		ts[i] = p.ctype.String()
	}
	return "callback:" + ret + "(" + strings.Join(ts, ",") + ")"
}

// trampoline returns the number of the trampoline (gopy_cb_<number>, see
// callbackABI.cbTrampolines) that calls a callback like cb, adding it if it
// is the first.
func (g *pyGen) trampoline(cb *callbackSig) int {
	for i, c := range g.cbs {
		if c.pyType() == cb.pyType() {
			return i
		}
	}
	g.cbs = append(g.cbs, cb)
	return len(g.cbs) - 1
}
