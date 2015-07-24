// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"golang.org/x/tools/go/types"
)

func needWrapType(typ types.Type) bool {
	switch typ.(type) {
	case *types.Struct:
		return true
	case *types.Named:
		switch typ.Underlying().(type) {
		case *types.Struct:
			return true
		}
	}
	return false
}

func cTypeName(typ types.Type) string {
	switch typ := typ.(type) {
	case *types.Basic:
		kind := typ.Kind()
		o, ok := typedescr[kind]
		if ok {
			return o.ctype
		}
	}
	return typ.String()
}

func cgoTypeName(typ types.Type) string {
	switch typ := typ.(type) {
	case *types.Basic:
		kind := typ.Kind()
		o, ok := typedescr[kind]
		if ok {
			return o.cgotype
		}
	}
	return typ.String()
}

type typedesc struct {
	ctype   string
	cgotype string
	pyfmt   string
}

var typedescr = map[types.BasicKind]typedesc{
	types.Bool: typedesc{
		ctype:   "_Bool",
		cgotype: "GoBool",
		pyfmt:   "b",
	},

	types.Int: typedesc{
		ctype:   "int",
		cgotype: "GoInt",
		pyfmt:   "i",
	},

	types.Int8: typedesc{
		ctype:   "int8_t",
		cgotype: "GoInt8",
		pyfmt:   "c",
	},

	types.Int16: typedesc{
		ctype:   "int16_t",
		cgotype: "GoInt16",
		pyfmt:   "h",
	},

	types.Int32: typedesc{
		ctype:   "int32_t",
		cgotype: "GoInt32",
		pyfmt:   "i",
	},

	types.Int64: typedesc{
		ctype:   "int64_t",
		cgotype: "GoInt64",
		pyfmt:   "k",
	},

	types.Uint: typedesc{
		ctype:   "unsigned int",
		cgotype: "GoUint",
		pyfmt:   "I",
	},

	types.Uint8: typedesc{
		ctype:   "uint8_t",
		cgotype: "GoUint8",
		pyfmt:   "b",
	},

	types.Uint16: typedesc{
		ctype:   "uint16_t",
		cgotype: "GoUint16",
		pyfmt:   "H",
	},

	types.Uint32: typedesc{
		ctype:   "uint32_t",
		cgotype: "GoUint32",
		pyfmt:   "I",
	},

	types.Uint64: typedesc{
		ctype:   "uint64_t",
		cgotype: "GoUint64",
		pyfmt:   "K",
	},

	types.Float32: typedesc{
		ctype:   "float",
		cgotype: "float",
		pyfmt:   "f",
	},

	types.Float64: typedesc{
		ctype:   "double",
		cgotype: "double",
		pyfmt:   "d",
	},

	types.Complex64: typedesc{
		ctype:   "float complex",
		cgotype: "GoComplex64",
		pyfmt:   "D",
	},

	types.Complex128: typedesc{
		ctype:   "double complex",
		cgotype: "GoComplex128",
		pyfmt:   "D",
	},

	types.String: typedesc{
		ctype:   "const char*",
		cgotype: "GoString",
		pyfmt:   "s",
	},

	types.UnsafePointer: typedesc{
		ctype:   "void*",
		cgotype: "void*",
		pyfmt:   "?",
	},
}
