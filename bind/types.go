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
	Obj() types.Object
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

func pyTypeName(typ types.Type) string {
	switch typ := typ.(type) {
	case *types.Basic:
		kind := typ.Kind()
		o, ok := typedescr[kind]
		if ok {
			return o.pysig
		}
	}
	return "object"
}

type typedesc struct {
	ctype   string
	cgotype string
	pyfmt   string
	pysig   string
}

var typedescr = map[types.BasicKind]typedesc{
	types.Bool: typedesc{
		ctype:   "_Bool",
		cgotype: "GoBool",
		pyfmt:   "b",
		pysig:   "bool",
	},

	types.Int: typedesc{
		ctype:   "int",
		cgotype: "GoInt",
		pyfmt:   "i",
		pysig:   "int",
	},

	types.Int8: typedesc{
		ctype:   "int8_t",
		cgotype: "GoInt8",
		pyfmt:   "c",
		pysig:   "int",
	},

	types.Int16: typedesc{
		ctype:   "int16_t",
		cgotype: "GoInt16",
		pyfmt:   "h",
		pysig:   "int",
	},

	types.Int32: typedesc{
		ctype:   "int32_t",
		cgotype: "GoInt32",
		pyfmt:   "i",
		pysig:   "long",
	},

	types.Int64: typedesc{
		ctype:   "int64_t",
		cgotype: "GoInt64",
		pyfmt:   "k",
		pysig:   "long",
	},

	types.Uint: typedesc{
		ctype:   "unsigned int",
		cgotype: "GoUint",
		pyfmt:   "I",
		pysig:   "int",
	},

	types.Uint8: typedesc{
		ctype:   "uint8_t",
		cgotype: "GoUint8",
		pyfmt:   "b",
		pysig:   "int",
	},

	types.Uint16: typedesc{
		ctype:   "uint16_t",
		cgotype: "GoUint16",
		pyfmt:   "H",
		pysig:   "int",
	},

	types.Uint32: typedesc{
		ctype:   "uint32_t",
		cgotype: "GoUint32",
		pyfmt:   "I",
		pysig:   "long",
	},

	types.Uint64: typedesc{
		ctype:   "uint64_t",
		cgotype: "GoUint64",
		pyfmt:   "K",
		pysig:   "long",
	},

	types.Float32: typedesc{
		ctype:   "float",
		cgotype: "float",
		pyfmt:   "f",
		pysig:   "float",
	},

	types.Float64: typedesc{
		ctype:   "double",
		cgotype: "double",
		pyfmt:   "d",
		pysig:   "float",
	},

	types.Complex64: typedesc{
		ctype:   "float complex",
		cgotype: "GoComplex64",
		pyfmt:   "D",
		pysig:   "float",
	},

	types.Complex128: typedesc{
		ctype:   "double complex",
		cgotype: "GoComplex128",
		pyfmt:   "D",
		pysig:   "float",
	},

	types.String: typedesc{
		ctype:   "const char*",
		cgotype: "GoString",
		pyfmt:   "s",
		pysig:   "str",
	},

	types.UnsafePointer: typedesc{
		ctype:   "void*",
		cgotype: "void*",
		pyfmt:   "?",
		pysig:   "object",
	},
}
