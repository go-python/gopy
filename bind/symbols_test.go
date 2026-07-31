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

// TestBuildTupleHoldsGIL guards against a regression class already fixed
// once for complex64/complex128 (see needsGILForArgMarshal in gen_func.go):
// generated code that touches a raw *C.PyObject must never run outside a
// PyGILState_Ensure/Release (or equivalent) bracket.
//
// buildTuple emits PyTuple_New/PyTuple_SetItem -- pure PyObject manipulation
// -- as a bare code fragment with no way to know whether its caller already
// holds the GIL. Its one current caller (addSignatureType's py2go closure)
// happens to bracket it correctly today, but that's a convention external
// to buildTuple, not something buildTuple enforces. This test requires
// buildTuple's own output to be self-bracketing, so a future caller can't
// reintroduce the complex64/128 bug by forgetting to hold the GIL.
func TestBuildTupleHoldsGIL(t *testing.T) {
	tuple := types.NewTuple(
		types.NewVar(token.NoPos, nil, "i", types.Typ[types.Int]),
		types.NewVar(token.NoPos, nil, "s", types.Typ[types.String]),
	)

	got, err := current.buildTuple(tuple, "_fcargs", "_fun_arg")
	if err != nil {
		t.Fatalf("buildTuple returned error: %v", err)
	}

	ensureCount := strings.Count(got, "C.PyGILState_Ensure()")
	releaseCount := strings.Count(got, "C.PyGILState_Release(")
	if ensureCount == 0 || releaseCount == 0 {
		t.Fatalf("buildTuple output does not self-bracket its PyTuple_* calls "+
			"with PyGILState_Ensure/Release; got:\n%s", got)
	}
	if ensureCount != releaseCount {
		t.Fatalf("mismatched PyGILState_Ensure/Release counts (%d vs %d) in "+
			"buildTuple output:\n%s", ensureCount, releaseCount, got)
	}

	ensureIdx := strings.Index(got, "C.PyGILState_Ensure()")
	firstTupleIdx := strings.Index(got, "C.PyTuple_")
	lastTupleIdx := strings.LastIndex(got, "C.PyTuple_")
	releaseIdx := strings.LastIndex(got, "C.PyGILState_Release(")

	if ensureIdx == -1 || firstTupleIdx == -1 || releaseIdx == -1 {
		t.Fatalf("could not locate expected markers in buildTuple output:\n%s", got)
	}
	if !(ensureIdx < firstTupleIdx && lastTupleIdx < releaseIdx) {
		t.Fatalf("PyTuple_* calls are not nested inside the "+
			"PyGILState_Ensure/Release bracket in buildTuple output:\n%s", got)
	}
}
