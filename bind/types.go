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
	switch typ.(type) {
	case *types.Struct:
		return true
	case *types.Named:
		switch typ.Underlying().(type) {
		case *types.Struct:
			return true
		}
	case *types.Array:
		return true
	case *types.Slice:
		return true
	case *types.Interface:
		return true
	}
	return false
}
