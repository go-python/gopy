// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/token"
	"go/types"
	"strings"
	"testing"
)

// TestAddSignatureTypeHoldsGILThroughoutClosure guards against the family of
// GIL bugs already fixed once for complex64/complex128 arguments (see
// needsGILForArgMarshal in gen_func.go): generated code that touches a raw
// *C.PyObject must never run outside a PyGILState_Ensure/Release bracket.
//
// addSignatureType's py2go closure is the one place this whole family shows
// up together: it builds a PyTuple from the Go-side arguments (PyTuple_New/
// SetItem, via buildTuple), invokes the Python callback
// (PyObject_CallObject), and converts the *C.PyObject result back to a Go
// value (pyObjectToGo, e.g. PyBytes_AsString) before decref'ing it. All of
// that touches raw PyObjects and must happen inside the single
// PyGILState_Ensure/Release the closure takes out -- not just the tuple
// build (buildTuple itself does not bracket its own output; it relies on
// its only caller, this closure, to hold the GIL across build, call, and
// return-conversion together). Getting the release point wrong here is not
// hypothetical: _examples/funcs' CallBackRval passes a bool-returning
// callback through exactly this path today.
func TestAddSignatureTypeHoldsGILThroughoutClosure(t *testing.T) {
	sig := types.NewSignature(nil,
		types.NewTuple(types.NewVar(token.NoPos, nil, "x", types.Typ[types.Int])),
		types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.String])),
		false)

	if err := current.addSignatureType(nil, nil, sig, 0, "sigtest_id", "sigtest"); err != nil {
		t.Fatalf("addSignatureType returned error: %v", err)
	}

	sym := current.symtype(sig)
	if sym == nil {
		t.Fatalf("addSignatureType did not register a symbol for %s", current.fullTypeString(sig))
	}
	got := sym.py2go

	ensureIdx := strings.Index(got, "C.PyGILState_Ensure()")
	lastTupleIdx := strings.LastIndex(got, "C.PyTuple_")
	callIdx := strings.Index(got, "C.PyObject_CallObject(")
	convIdx := strings.Index(got, "C.PyBytes_AsString(_fcret)")
	decrefIdx := strings.Index(got, "C.gopy_decref(_fcret)")
	releaseIdx := strings.LastIndex(got, "C.PyGILState_Release(_gstate)")

	for name, idx := range map[string]int{
		"PyGILState_Ensure":        ensureIdx,
		"PyTuple_*":                lastTupleIdx,
		"PyObject_CallObject":      callIdx,
		"PyBytes_AsString":         convIdx,
		"gopy_decref(_fcret)":      decrefIdx,
		"final PyGILState_Release": releaseIdx,
	} {
		if idx == -1 {
			t.Fatalf("could not locate expected %s call in generated closure:\n%s", name, got)
		}
	}

	// Everything that touches a PyObject -- building the tuple, invoking
	// the callback, converting and decref'ing the result -- must occur
	// strictly between the Ensure and the final Release.
	if !(ensureIdx < lastTupleIdx && lastTupleIdx < callIdx && callIdx < convIdx &&
		convIdx < releaseIdx && decrefIdx < releaseIdx) {
		t.Fatalf("generated closure does not hold the GIL across the entire "+
			"build-tuple/call/convert/decref sequence:\n%s", got)
	}
}
