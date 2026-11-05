// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import "testing"

// TestArgCallExprPrecedenceMatchesPremarshalDecision guards against the
// failure mode flagged in the PR #401 review of needsGILForArgMarshal: the
// call-argument expression for a Python-wrapped argument used to be decided
// by a switch with a fixed case order (ifchandle&&interface{} > isSignature
// > premarshalled > py2go > default), while a *separate* loop decided,
// independently, whether to emit a "_premarshalN := ..." declaration before
// the GIL is released. Nothing today makes an argument match both an
// earlier switch case and needsGILForArgMarshal, but if it ever did, the
// premarshal declaration would be emitted and never referenced --
// "_premarshalN declared and not used" -- a compile error surfacing in the
// user's generated package, not in gopy.
//
// argCallExpr is the fix: both the premarshal pass and the call-argument
// pass now call this single function with the same inputs, so a
// premarshalled variable is declared if and only if it is also the
// expression that gets used. This test exercises that precedence directly,
// including a contrived interface{}/*C.PyObject collision that cannot
// happen with today's type table (interface{}'s cgoname is "*C.char") but
// is exactly the shape the review warned about.
func TestArgCallExprPrecedenceMatchesPremarshalDecision(t *testing.T) {
	complex128Sym := &symbol{
		goname:       "complex128",
		cgoname:      "*C.PyObject",
		py2go:        "complex128FromPyObject",
		py2goParenEx: "",
	}

	ifaceHandleSym := &symbol{
		goname:       "interface{}",
		cgoname:      "*C.PyObject", // contrived: not true of any symbol today
		py2go:        "wouldBeMarshalFunc",
		py2goParenEx: "",
	}

	sigSym := &symbol{
		kind:         skSignature,
		goname:       "func(int) string",
		cgoname:      "*C.PyObject",
		py2go:        "func (x int) string { ... }",
		py2goParenEx: "",
	}

	plainStringSym := &symbol{
		goname:       "string",
		cgoname:      "*C.char",
		py2go:        "C.GoString",
		py2goParenEx: "",
	}

	tests := []struct {
		name         string
		ifchandle    bool
		isVariadic   bool
		sym          *symbol
		wantExpr     string
		wantNeedsGIL bool
	}{
		{
			name:         "complex128 arg is premarshalled",
			sym:          complex128Sym,
			wantExpr:     "complex128FromPyObject(x)",
			wantNeedsGIL: true,
		},
		{
			name:         "complex128 arg in variadic tail is not premarshalled",
			isVariadic:   true,
			sym:          complex128Sym,
			wantExpr:     "complex128FromPyObject(x)",
			wantNeedsGIL: false,
		},
		{
			name:         "ifchandle interface{} wins over premarshalling even if cgoname collides",
			ifchandle:    true,
			sym:          ifaceHandleSym,
			wantExpr:     `gopyh.VarFromHandle((gopyh.CGoHandle)(x), "interface{}")`,
			wantNeedsGIL: false,
		},
		{
			name:         "isSignature wins over premarshalling",
			sym:          sigSym,
			wantExpr:     "func (x int) string { ... }",
			wantNeedsGIL: false,
		},
		{
			name:         "plain string arg is not premarshalled",
			sym:          plainStringSym,
			wantExpr:     "C.GoString(x)",
			wantNeedsGIL: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExpr, gotNeedsGIL := argCallExpr(tt.ifchandle, tt.isVariadic, "x", tt.sym)
			if gotExpr != tt.wantExpr {
				t.Errorf("argCallExpr() expr = %q, want %q", gotExpr, tt.wantExpr)
			}
			if gotNeedsGIL != tt.wantNeedsGIL {
				t.Errorf("argCallExpr() needsGIL = %v, want %v", gotNeedsGIL, tt.wantNeedsGIL)
			}
		})
	}
}
