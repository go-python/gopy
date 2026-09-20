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
// Only what is needed so far is supported: no results, and parameters
// limited to pointer/interface handles, numbers, bool and string.  Any other
// callback type makes the function be skipped, see genFuncSig.

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

type cffiCallback struct {
	params []cffiCBParam
}

// cffiCallback returns how to pass sym, a func-typed argument, to Go
// as a callback, or nil if its type isn't supported.
func (g *pyGen) cffiCallback(sym *symbol) *cffiCallback {
	sig, ok := sym.GoType().Underlying().(*types.Signature)
	if !ok || sig.Results().Len() != 0 || sig.Variadic() {
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
	return cb
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
	bt, ok := typ.Underlying().(*types.Basic)
	if !ok {
		return p, false
	}
	switch k := bt.Kind(); {
	case types.Int <= k && k <= types.Int64:
		p.ctype, p.conv = "int64_t", fmt.Sprintf("C.int64_t(%s)", nm)
	case types.Uint <= k && k <= types.Uintptr:
		p.ctype, p.conv = "uint64_t", fmt.Sprintf("C.uint64_t(%s)", nm)
	case k == types.Float32 || k == types.Float64:
		p.ctype, p.conv = "double", fmt.Sprintf("C.double(%s)", nm)
	case k == types.Bool:
		p.ctype, p.conv = "bool", fmt.Sprintf("C.uint8_t(boolGoToPy(bool(%s)))", nm)
	case k == types.String:
		// the python side copies it, so it can be freed once the call returns
		p.ctype, p.conv = "char*", "_c"+nm
		p.pre = fmt.Sprintf("_c%[1]s := C.CString(string(%[1]s))\ndefer C.free(unsafe.Pointer(_c%[1]s))\n", nm)
	default:
		return p, false
	}
	return p, true
}

func (cb *cffiCallback) ctypes() []string {
	ts := make([]string, len(cb.params))
	for i, p := range cb.params {
		ts[i] = p.ctype
	}
	return ts
}

// pyType is the type given to the callback's parameter in build.py, which
// callback_setup in cffi_build.py takes apart.
func (cb *cffiCallback) pyType() string {
	return "callback:void(" + strings.Join(cb.ctypes(), ",") + ")"
}

// cffiCallbackPrologue returns the Go statements that set up the callback
// argument named anm, ahead of cffiCallbackLit.
func cffiCallbackPrologue(anm string) string {
	return fmt.Sprintf("_cbfp_%[1]s := %[1]s\n_cbs_%[1]s := new(gopyCallbackScope)\ndefer _cbs_%[1]s.close()\n", anm)
}

// cffiCallbackLit returns a Go func literal that calls the Python callable
// passed as the argument named anm.
func (g *pyGen) cffiCallbackLit(cb *cffiCallback, anm string) string {
	var decl, pre, args []string
	args = append(args, "_cbfp_"+anm)
	for _, p := range cb.params {
		decl = append(decl, p.name+" "+p.gotyp)
		if p.pre != "" {
			pre = append(pre, p.pre)
		}
		args = append(args, p.conv)
	}
	return fmt.Sprintf("func(%s) {\nif !_cbs_%s.enter() {\nreturn\n}\ndefer _cbs_%[2]s.leave()\n%sC.gopy_cb_%d(%s)\n}",
		strings.Join(decl, ", "), anm, strings.Join(pre, ""), g.cffiTrampoline(cb), strings.Join(args, ", "))
}

// cffiTrampoline returns the number of the C trampoline that calls a callback
// like cb, adding it if it is the first.
func (g *pyGen) cffiTrampoline(cb *cffiCallback) int {
	key := strings.Join(cb.ctypes(), ",")
	for i, k := range g.cbSigs {
		if k == key {
			return i
		}
	}
	g.cbSigs = append(g.cbSigs, key)
	return len(g.cbSigs) - 1
}

// spliceCFFITrampolines writes the trampolines into the cgo preamble.
func (g *pyGen) spliceCFFITrampolines() {
	if !g.isCFFI() {
		return
	}
	var c strings.Builder
	for i, key := range g.cbSigs {
		params := []string{"void* f"}
		var args, ptypes []string
		if key != "" {
			for j, t := range strings.Split(key, ",") {
				if t == "bool" {
					t = "uint8_t"
				}
				params = append(params, fmt.Sprintf("%s a%d", t, j))
				args = append(args, fmt.Sprintf("a%d", j))
				ptypes = append(ptypes, t)
			}
		} else {
			ptypes = append(ptypes, "void")
		}
		fmt.Fprintf(&c, "static inline void gopy_cb_%d(%s) { ((void (*)(%s))f)(%s); }\n",
			i, strings.Join(params, ", "), strings.Join(ptypes, ", "), strings.Join(args, ", "))
	}
	b := bytes.Replace(g.gofile.buf.Bytes(), []byte(cffiTrampolinesKey), []byte(c.String()), 1)
	g.gofile.buf = bytes.NewBuffer(b)
}
