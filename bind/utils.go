// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
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
