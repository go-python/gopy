// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"golang.org/x/tools/go/types"
)

type Object interface {
	Package() *Package
	ID() string
	Doc() string
	GoName() string
}

type Type interface {
	Object
	GoType() types.Type
}

func needWrapType(typ types.Type) bool {
	switch typ := typ.(type) {
	case *types.Basic:
		return false
	case *types.Struct:
		return true
	case *types.Named:
		switch ut := typ.Underlying().(type) {
		case *types.Basic:
			return false
		default:
			return needWrapType(ut)
		}
	case *types.Array:
		return true
	case *types.Slice:
		return true
	case *types.Interface:
		wrap := true
		if typ.Underlying() == universe.syms["error"].GoType().Underlying() {
			wrap = false
		}
		return wrap
	case *types.Signature:
		return true
	case *types.Pointer:
		return needWrapType(typ.Elem())
	}
	return false
}
