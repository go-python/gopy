// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/types"
	"reflect"
)

// goPackage is the fake package that contains all our standard slice / map
// types that we export
var goPackage *Package

// makeGoPackage
func makeGoPackage() {
	gopk := types.NewPackage("go", "go")
	goPackage = &Package{pkg: gopk, syms: universe, objs: map[string]Object{}}
	Packages = append(Packages, goPackage)
}

// addStdSliceMaps adds std Slice and Map types to universe
func addStdSliceMaps() {
	makeGoPackage()
	gopk := goPackage.pkg
	sltyps := []string{"int", "int64", "int32", "int16", "int8", "uint", "uint64", "uint32", "uint16", "uint8", "bool", "byte", "rune", "float64", "float32", "string"}
	for _, tn := range sltyps {
		universe.addSliceType(gopk, nil, types.NewSlice(universe.sym(tn).gotyp), skType, "Slice_"+tn, "[]"+tn)
	}
}

// stdBasicTypes returns the basic int, float etc types as symbols
func stdBasicTypes() map[string]*symbol {
	look := types.Universe.Lookup
	syms := map[string]*symbol{
		"bool": {
			gopkg:   look("bool").Pkg(),
			goobj:   look("bool"),
			gotyp:   look("bool").Type(),
			kind:    skType | skBasic,
			goname:  "bool",
			id:      "bool",
			cgoname: "C.char",
			cpyname: "bool",
			pysig:   "bool",
			go2py:   "boolGoToPy",
			py2go:   "boolPyToGo",
			zval:    "false",
			pyfmt:   "", // TODO:
		},

		"byte": {
			gopkg:   look("byte").Pkg(),
			goobj:   look("byte"),
			gotyp:   look("byte").Type(),
			kind:    skType | skBasic,
			goname:  "byte",
			id:      "int",
			cpyname: "uint8_t",
			cgoname: "C.char",
			pysig:   "int", // FIXME(sbinet) py2/py3
			go2py:   "C.char",
			py2go:   "byte",
			zval:    "0",
			pyfmt:   "b",
		},

		"int": {
			gopkg:   look("int").Pkg(),
			goobj:   look("int"),
			gotyp:   look("int").Type(),
			kind:    skType | skBasic,
			goname:  "int",
			id:      "int",
			cpyname: "int",
			cgoname: "C.long", // see below for 64 bit version
			pysig:   "int",
			go2py:   "C.long",
			py2go:   "int",
			zval:    "0",
			pyfmt:   "i",
		},

		"int8": {
			gopkg:   look("int8").Pkg(),
			goobj:   look("int8"),
			gotyp:   look("int8").Type(),
			kind:    skType | skBasic,
			goname:  "int8",
			id:      "int",
			cpyname: "int8_t",
			cgoname: "C.char",
			pysig:   "int",
			go2py:   "C.char",
			py2go:   "int8",
			zval:    "0",
			pyfmt:   "b",
		},

		"int16": {
			gopkg:   look("int16").Pkg(),
			goobj:   look("int16"),
			gotyp:   look("int16").Type(),
			kind:    skType | skBasic,
			goname:  "int16",
			id:      "int16",
			cpyname: "int16_t",
			cgoname: "C.short",
			pysig:   "int",
			go2py:   "C.short",
			py2go:   "int16",
			zval:    "0",
			pyfmt:   "h",
		},

		"int32": {
			gopkg:   look("int32").Pkg(),
			goobj:   look("int32"),
			gotyp:   look("int32").Type(),
			kind:    skType | skBasic,
			goname:  "int32",
			id:      "int",
			cpyname: "int32_t",
			cgoname: "C.long",
			pysig:   "int",
			go2py:   "C.long",
			py2go:   "int32",
			zval:    "0",
			pyfmt:   "i",
		},

		"int64": {
			gopkg:   look("int64").Pkg(),
			goobj:   look("int64"),
			gotyp:   look("int64").Type(),
			kind:    skType | skBasic,
			goname:  "int64",
			id:      "int64",
			cpyname: "int64_t",
			cgoname: "C.longlong",
			pysig:   "long",
			go2py:   "C.longlong",
			py2go:   "int64",
			zval:    "0",
			pyfmt:   "k",
		},

		"uint": {
			gopkg:   look("uint").Pkg(),
			goobj:   look("uint"),
			gotyp:   look("uint").Type(),
			kind:    skType | skBasic,
			goname:  "uint",
			id:      "uint",
			cpyname: "unsigned int",
			cgoname: "C.uint",
			pysig:   "int",
			go2py:   "C.uint",
			py2go:   "uint",
			zval:    "0",
			pyfmt:   "I",
		},

		"uint8": {
			gopkg:   look("uint8").Pkg(),
			goobj:   look("uint8"),
			gotyp:   look("uint8").Type(),
			kind:    skType | skBasic,
			goname:  "uint8",
			id:      "uint8",
			cpyname: "uint8_t",
			cgoname: "C.uchar",
			pysig:   "int",
			go2py:   "C.uchar",
			py2go:   "uint8",
			zval:    "0",
			pyfmt:   "B",
		},

		"uint16": {
			gopkg:   look("uint16").Pkg(),
			goobj:   look("uint16"),
			gotyp:   look("uint16").Type(),
			kind:    skType | skBasic,
			goname:  "uint16",
			id:      "uint16",
			cpyname: "uint16_t",
			cgoname: "C.ushort",
			pysig:   "int",
			go2py:   "C.ushort",
			py2go:   "uint16",
			zval:    "0",
			pyfmt:   "H",
		},

		"uint32": {
			gopkg:   look("uint32").Pkg(),
			goobj:   look("uint32"),
			gotyp:   look("uint32").Type(),
			kind:    skType | skBasic,
			goname:  "uint32",
			id:      "uint32",
			cpyname: "uint32_t",
			cgoname: "C.ulong",
			pysig:   "long",
			go2py:   "C.ulong",
			py2go:   "uint32",
			zval:    "0",
			pyfmt:   "I",
		},

		"uint64": {
			gopkg:   look("uint64").Pkg(),
			goobj:   look("uint64"),
			gotyp:   look("uint64").Type(),
			kind:    skType | skBasic,
			goname:  "uint64",
			id:      "uint64",
			cpyname: "uint64_t",
			cgoname: "C.ulonglong",
			pysig:   "long",
			go2py:   "C.ulonglong",
			py2go:   "uint64",
			zval:    "0",
			pyfmt:   "K",
		},

		"uintptr": {
			gopkg:   look("uintptr").Pkg(),
			goobj:   look("uintptr"),
			gotyp:   look("uintptr").Type(),
			kind:    skType | skBasic,
			goname:  "uintptr",
			id:      "uintptr",
			cpyname: "uint64_t",
			cgoname: "C.ulonglong",
			pysig:   "long",
			go2py:   "C.ulonglong",
			py2go:   "uintptr",
			zval:    "0",
			pyfmt:   "K",
		},

		"float32": {
			gopkg:   look("float32").Pkg(),
			goobj:   look("float32"),
			gotyp:   look("float32").Type(),
			kind:    skType | skBasic,
			goname:  "float32",
			id:      "float32",
			cpyname: "float",
			cgoname: "C.float",
			pysig:   "float",
			go2py:   "C.float",
			py2go:   "float32",
			zval:    "0",
			pyfmt:   "f",
		},

		"float64": {
			gopkg:   look("float64").Pkg(),
			goobj:   look("float64"),
			gotyp:   look("float64").Type(),
			kind:    skType | skBasic,
			goname:  "float64",
			id:      "float64",
			cpyname: "double",
			cgoname: "C.double",
			pysig:   "float",
			go2py:   "C.double",
			py2go:   "float64",
			zval:    "0",
			pyfmt:   "d",
		},

		"complex64": {
			gopkg:   look("complex64").Pkg(),
			goobj:   look("complex64"),
			gotyp:   look("complex64").Type(),
			kind:    skType | skBasic,
			goname:  "complex64",
			id:      "complex64",
			cpyname: "PyObject*",
			cgoname: "*C.PyObject",
			pysig:   "complex",
			go2py:   "complex64GoToPy",
			py2go:   "complex64PyToGo",
			zval:    "0",
			pyfmt:   "O&",
		},

		"complex128": {
			gopkg:   look("complex128").Pkg(),
			goobj:   look("complex128"),
			gotyp:   look("complex128").Type(),
			kind:    skType | skBasic,
			goname:  "complex128",
			id:      "complex128",
			cpyname: "PyObject*",
			cgoname: "*C.PyObject",
			pysig:   "complex",
			go2py:   "complex128GoToPy",
			py2go:   "complex128PyToGo",
			zval:    "0",
			pyfmt:   "O&",
		},

		"string": {
			gopkg:   look("string").Pkg(),
			goobj:   look("string"),
			gotyp:   look("string").Type(),
			kind:    skType | skBasic,
			goname:  "string",
			id:      "string",
			cpyname: "char*",
			cgoname: "*C.char",
			pysig:   "str",
			go2py:   "C.CString",
			py2go:   "C.GoString",
			zval:    `""`,
			pyfmt:   "s",
		},

		"rune": { // FIXME(sbinet) py2/py3
			gopkg:   look("rune").Pkg(),
			goobj:   look("rune"),
			gotyp:   look("rune").Type(),
			kind:    skType | skBasic,
			goname:  "rune",
			id:      "rune",
			cpyname: "int32_t",
			cgoname: "C.long",
			pysig:   "str",
			go2py:   "C.long",
			py2go:   "rune",
			zval:    "0",
			pyfmt:   "i",
		},

		"error": {
			gopkg:        look("error").Pkg(),
			goobj:        look("error"),
			gotyp:        look("error").Type(),
			kind:         skType | skInterface,
			goname:       "error",
			id:           "error",
			cpyname:      "char*",
			cgoname:      "*C.char",
			pysig:        "str",
			go2py:        "C.CString",
			py2go:        "errors.New(C.GoString",
			py2goParenEx: ")",
			zval:         `""`,
			pyfmt:        "O&",
		},
	}

	if reflect.TypeOf(int(0)).Size() == 8 {
		syms["int"] = &symbol{
			gopkg:   look("int").Pkg(),
			goobj:   look("int"),
			gotyp:   look("int").Type(),
			kind:    skType | skBasic,
			goname:  "int",
			id:      "int",
			cpyname: "int64_t",
			cgoname: "C.longlong",
			pysig:   "int",
			go2py:   "C.longlong",
			py2go:   "int",
			zval:    "0",
			pyfmt:   "k",
		}
		syms["uint"] = &symbol{
			gopkg:   look("uint").Pkg(),
			goobj:   look("uint"),
			gotyp:   look("uint").Type(),
			kind:    skType | skBasic,
			goname:  "uint",
			id:      "uint",
			cpyname: "uint64_t",
			cgoname: "C.ulonglong",
			pysig:   "int",
			go2py:   "C.ulonglong",
			py2go:   "uint",
			zval:    "0",
			pyfmt:   "K",
		}
	}

	// these are defined in: https://godoc.org/go/types
	for _, o := range []struct {
		kind  types.BasicKind
		tname string
		uname string
	}{
		{types.UntypedBool, "bool", "bool"},
		{types.UntypedInt, "int", "int"},
		{types.UntypedRune, "rune", "rune"},
		{types.UntypedFloat, "float64", "float"},
		{types.UntypedComplex, "complex128", "complex"},
		{types.UntypedString, "string", "string"},
		//FIXME(sbinet): what should be the python equivalent?
		//{types.UntypedNil, "nil", "nil"},
	} {
		sym := *syms[o.tname]
		n := "untyped " + o.uname
		syms[n] = &sym
	}

	return syms
}
