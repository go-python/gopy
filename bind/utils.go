// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"

	"golang.org/x/tools/go/types"
)

func isErrorType(typ types.Type) bool {
	return typ == types.Universe.Lookup("error").Type()
}

func isStringer(obj types.Object) bool {
	switch obj := obj.(type) {
	case *types.Func:
		if obj.Name() != "String" {
			return false
		}
		sig, ok := obj.Type().(*types.Signature)
		if !ok {
			return false
		}
		if sig.Recv() == nil {
			return false
		}
		if sig.Params().Len() != 0 {
			return false
		}
		res := sig.Results()
		if res.Len() != 1 {
			return false
		}
		ret := res.At(0).Type()
		if ret != types.Universe.Lookup("string").Type() {
			return false
		}
		return true
	default:
		return false
	}

	return false
}

func hasError(sig *types.Signature) bool {
	res := sig.Results()
	if res == nil || res.Len() <= 0 {
		return false
	}

	nerr := 0
	for i := 0; i < res.Len(); i++ {
		ret := res.At(i)
		if isErrorType(ret.Type()) {
			nerr++
		}
	}

	switch {
	case nerr == 0:
		return false
	case nerr == 1:
		return true
	default:
		panic(fmt.Errorf(
			"gopy: invalid number of comma-errors (%d)",
			nerr,
		))
	}

	return false
}

func isConstructor(sig *types.Signature) bool {
	//TODO(sbinet)
	return false
}
