// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"fmt"
	"go/types"
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
// Supported: parameters that are pointer/interface handles, numbers, bool,
// string or interface{} (which arrives as a string, as with pybindgen), and a
// result that is a number or bool.  Any other callback type makes the
// function be skipped, see genFuncSig.

// cffiTrampolinesKey stands in for the trampolines in the cgo preamble,
// which is written before the callback types that need them are known.
const cffiTrampolinesKey = "@@GOPY_CFFI_TRAMPOLINES@@"

// cffiCBParam is one parameter of a callback.
type cffiCBParam struct {
	name  string // name of the parameter in the func literal
	gotyp string // its Go type
	ctype string // how it crosses: int64_t, uint64_t, double, char* or bool
	pre   string // Go statements to run before the call, if any
	conv  string // Go expression converting it to ctype
}

// cffiCBResult is the result of a callback.
type cffiCBResult struct {
	gotyp string // its Go type
	ctype string // how it crosses: int64_t, uint64_t, double or bool
}

type cffiCallback struct {
	params []cffiCBParam
	ret    *cffiCBResult // nil if the callback has no result
}

// cffiBasicCType returns how a value of a basic Go type crosses, or "" if it
// can't.
func cffiBasicCType(k types.BasicKind) string {
	switch {
	case types.Int <= k && k <= types.Int64:
		return "int64_t"
	case types.Uint <= k && k <= types.Uintptr:
		return "uint64_t"
	case k == types.Float32 || k == types.Float64:
		return "double"
	case k == types.Bool:
		return "bool"
	case k == types.String:
		return "char*"
	}
	return ""
}

// cffiCType returns the C type used for ctype in C code: bool is a byte.
func cffiCType(ctype string) string {
	if ctype == "bool" {
		return "uint8_t"
	}
	return ctype
}

// cffiCallback returns how to pass sym, a func-typed argument, to Go
// as a callback, or nil if its type isn't supported.
func (g *pyGen) cffiCallback(sym *symbol) *cffiCallback {
	sig, ok := sym.GoType().Underlying().(*types.Signature)
	if !ok || sig.Results().Len() > 1 || sig.Variadic() {
		return nil
	}
	cb := &cffiCallback{}
	for i := 0; i < sig.Params().Len(); i++ {
		p, ok := cffiCallbackParam(sig.Params().At(i), i)
		if !ok {
			return nil
		}
		cb.params = append(cb.params, p)
	}
	if sig.Results().Len() == 1 {
		r, ok := cffiCallbackResult(sig.Results().At(0).Type())
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

func cffiCallbackParam(v *types.Var, i int) (cffiCBParam, bool) {
	typ := v.Type()
	vsym := current.symtype(typ)
	if vsym == nil {
		return cffiCBParam{}, false
	}
	nm := pySafeArg(v.Name(), i)
	p := cffiCBParam{name: nm, gotyp: current.typeGoName(typ)}

	if vsym.hasHandle() && vsym.isPtrOrIface() {
		p.ctype = "int64_t"
		p.conv = fmt.Sprintf("C.int64_t(%s(%s)%s)", vsym.go2py, nm, vsym.go2pyParenEx)
		return p, true
	}
	if vsym.goname == "interface{}" {
		p.ctype, p.conv = "char*", "_c"+nm
		p.pre = cStringPre(nm, fmt.Sprintf("%s(%s)%s", vsym.go2py, nm, vsym.go2pyParenEx))
		return p, true
	}
	bt, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return p, false
	}
	switch p.ctype = cffiBasicCType(bt.Kind()); p.ctype {
	case "":
		return p, false
	case "bool":
		p.conv = fmt.Sprintf("C.uint8_t(boolGoToPy(bool(%s)))", nm)
	case "char*":
		p.conv = "_c" + nm
		p.pre = cStringPre(nm, fmt.Sprintf("C.CString(string(%s))", nm))
	default:
		p.conv = fmt.Sprintf("C.%s(%s)", p.ctype, nm)
	}
	return p, true
}

func cffiCallbackResult(typ types.Type) (*cffiCBResult, bool) {
	bt, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return nil, false
	}
	switch ctype := cffiBasicCType(bt.Kind()); ctype {
	case "", "char*": // the python side has no way to give Go a string it owns
		return nil, false
	default:
		return &cffiCBResult{gotyp: current.typeGoName(typ), ctype: ctype}, true
	}
}

// pyType is the type given to the callback's parameter in build.py, which
// callback_setup in cffi_build.py takes apart:
// callback:<result or void>(<parameter types>)
func (cb *cffiCallback) pyType() string {
	ret := "void"
	if cb.ret != nil {
		ret = cb.ret.ctype
	}
	ts := make([]string, len(cb.params))
	for i, p := range cb.params {
		ts[i] = p.ctype
	}
	return "callback:" + ret + "(" + strings.Join(ts, ",") + ")"
}

// cffiCallbackPrologue returns the Go statements that set up the callback
// argument named anm, ahead of cffiCallbackLit.
func cffiCallbackPrologue(anm string) string {
	return fmt.Sprintf("_cbfp_%[1]s := %[1]s\n_cbs_%[1]s := new(gopyCallbackScope)\ndefer _cbs_%[1]s.close()\n", anm)
}

// cffiCallbackLit returns a Go func literal that calls the Python callable
// passed as the argument named anm.
func (g *pyGen) cffiCallbackLit(cb *cffiCallback, anm string) string {
	var decl, pre []string
	args := []string{"_cbfp_" + anm}
	for _, p := range cb.params {
		decl = append(decl, p.name+" "+p.gotyp)
		pre = append(pre, p.pre)
		args = append(args, p.conv)
	}
	call := fmt.Sprintf("C.gopy_cb_%d(%s)", g.cffiTrampoline(cb), strings.Join(args, ", "))
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

// cffiTrampoline returns the number of the C trampoline that calls a callback
// like cb, adding it if it is the first.
func (g *pyGen) cffiTrampoline(cb *cffiCallback) int {
	for i, c := range g.cbs {
		if c.pyType() == cb.pyType() {
			return i
		}
	}
	g.cbs = append(g.cbs, cb)
	return len(g.cbs) - 1
}

// spliceCFFITrampolines writes the trampolines into the cgo preamble.
func (g *pyGen) spliceCFFITrampolines() {
	if !g.isCFFI() {
		return
	}
	var c strings.Builder
	for i, cb := range g.cbs {
		params := []string{"void* f"}
		ptypes := []string{}
		args := []string{}
		for j, p := range cb.params {
			t := cffiCType(p.ctype)
			params = append(params, fmt.Sprintf("%s a%d", t, j))
			ptypes = append(ptypes, t)
			args = append(args, fmt.Sprintf("a%d", j))
		}
		if len(ptypes) == 0 {
			ptypes = append(ptypes, "void")
		}
		ret, retStmt := "void", ""
		if cb.ret != nil {
			ret, retStmt = cffiCType(cb.ret.ctype), "return "
		}
		fmt.Fprintf(&c, "static inline %s gopy_cb_%d(%s) { %s((%s (*)(%s))f)(%s); }\n",
			ret, i, strings.Join(params, ", "), retStmt, ret, strings.Join(ptypes, ", "), strings.Join(args, ", "))
	}
	b := bytes.Replace(g.gofile.buf.Bytes(), []byte(cffiTrampolinesKey), []byte(c.String()), 1)
	g.gofile.buf = bytes.NewBuffer(b)
}
