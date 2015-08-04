// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/tools/go/types"
)

var (
	universe *symtab
)

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("0x%d", h.Sum32())
}

// symkind describes the kinds of symbol
type symkind int

const (
	skConst symkind = 1 << iota
	skVar
	skFunc
	skType
)

// symbol is an exported symbol in a go package
type symbol struct {
	kind    symkind
	goobj   types.Object
	doc     string
	id      string
	goname  string
	cgoname string
	cpyname string

	// for types only

	pyfmt string
	pysig string
	c2py  string
	py2c  string
}

func (s symbol) isType() bool {
	return s.kind == skType
}

func (s symbol) isBasic() bool {
	_, ok := s.goobj.Type().(*types.Basic)
	return ok
}

func (s symbol) isArray() bool {
	_, ok := s.goobj.Type().(*types.Array)
	return ok
}

func (s symbol) isSlice() bool {
	_, ok := s.goobj.Type().(*types.Slice)
	return ok
}

func (s symbol) isStruct() bool {
	typ, ok := s.goobj.(*types.TypeName)
	if !ok {
		return false
	}
	_, ok = typ.Type().Underlying().(*types.Struct)
	return ok
}

func (s symbol) hasConverter() bool {
	return s.c2py != "" || s.py2c != ""
}

func (s symbol) cgotypename() string {
	typ := s.goobj.Type()
	switch typ := typ.(type) {
	case *types.Basic:
		n := typ.Name()
		if strings.HasPrefix(n, "untyped ") {
			n = string(n[len("untyped "):])
		}
		return n
	case *types.Named:
		obj := s.goobj
		switch typ.Underlying().(type) {
		case *types.Struct:
			return s.cgoname
		case *types.Interface:
			if obj.Name() == "error" {
				return "error"
			}
		}
	}
	return s.cgoname
}

// symtab is a table of symbols in a go package
type symtab struct {
	pkg    *types.Package
	syms   map[string]*symbol
	parent *symtab
}

func newSymtab(pkg *types.Package, parent *symtab) *symtab {
	if parent == nil {
		parent = universe
	}
	s := &symtab{
		pkg:    pkg,
		syms:   make(map[string]*symbol),
		parent: parent,
	}
	return s
}

func (sym *symtab) names() []string {
	names := make([]string, 0, len(sym.syms))
	for n := range sym.syms {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (sym *symtab) sym(n string) *symbol {
	s, ok := sym.syms[n]
	if ok {
		return s
	}
	if sym.parent != nil {
		return sym.parent.sym(n)
	}
	return nil
}

func (sym *symtab) typeof(n string) *symbol {
	s := sym.sym(n)
	switch s.kind {
	case skVar, skConst:
		tname := sym.typename(s.goobj.Type())
		return sym.sym(tname)
	case skFunc:
		//FIXME(sbinet): really?
		return s
	case skType:
		return s
	default:
		panic(fmt.Errorf("unhandled symbol kind (%v)", s.kind))
	}
	panic("unreachable")
}

func (sym *symtab) typename(t types.Type) string {
	return types.TypeString(t, types.RelativeTo(sym.pkg))
}

func (sym *symtab) symtype(t types.Type) *symbol {
	tname := sym.typename(t)
	s := sym.sym(tname)
	if s != nil {
		return s
	}
	switch typ := t.(type) {
	case *types.Pointer:
		s = sym.symtype(typ.Elem())
		if s == nil {
			return nil
		}
		sym.addType(s.goobj, typ)
	}
	return sym.sym(tname)
}

func (sym *symtab) addSymbol(obj types.Object) {
	n := obj.Name()
	pkg := obj.Pkg()
	id := n
	if pkg != nil {
		id = pkg.Name() + "_" + n
	}
	switch obj.(type) {
	case *types.Const:
		sym.syms[n] = &symbol{
			goobj:   obj,
			kind:    skConst,
			id:      id,
			goname:  n,
			cgoname: "cgo_const_" + id,
			cpyname: "cpy_const_" + id,
		}
		sym.addType(obj, obj.Type())

	case *types.Var:
		sym.syms[n] = &symbol{
			goobj:   obj,
			kind:    skVar,
			id:      id,
			goname:  n,
			cgoname: "cgo_var_" + id,
			cpyname: "cpy_var_" + id,
		}
		sym.addType(obj, obj.Type())

	case *types.Func:
		sym.syms[n] = &symbol{
			goobj:   obj,
			kind:    skFunc,
			id:      id,
			goname:  n,
			cgoname: "cgo_func_" + id,
			cpyname: "cpy_func_" + id,
		}

	case *types.TypeName:
		sym.addType(obj, obj.Type())
	}
}

func (sym *symtab) addType(obj types.Object, t types.Type) {
	n := sym.typename(t)
	pkg := obj.Pkg()
	id := n
	if pkg != nil {
		id = pkg.Name() + "_" + n
	}
	kind := skType
	switch typ := t.(type) {
	case *types.Basic:
		styp := sym.symtype(typ)
		if styp == nil {
			panic(fmt.Errorf("builtin type not already known [%s]!", n))
		}

	case *types.Array:
		enam := sym.typename(typ.Elem())
		elt := sym.sym(enam)
		if elt.goname == "" {
			eobj := sym.pkg.Scope().Lookup(enam)
			if eobj == nil {
				panic(fmt.Errorf("could not look-up %q!\n", enam))
			}
			sym.addSymbol(eobj)
			elt = sym.typeof(enam)
		}
		id := hash(id)
		sym.syms[n] = &symbol{
			goobj:   obj,
			kind:    kind,
			id:      id,
			goname:  n,
			cgoname: "cgo_type_" + id,
			cpyname: "cpy_type_" + id,
			pyfmt:   "O&",
			pysig:   "[]" + elt.pysig,
			c2py:    "cgopy_cnv_c2py_" + id,
			py2c:    "cgopy_cnv_py2c_" + id,
		}

	case *types.Slice:
		enam := sym.typename(typ.Elem())
		elt := sym.sym(enam)
		if elt.goname == "" {
			eobj := sym.pkg.Scope().Lookup(enam)
			if eobj == nil {
				panic(fmt.Errorf("could not look-up %q!\n", enam))
			}
			sym.addSymbol(eobj)
			elt = sym.typeof(enam)
		}
		id := hash(id)
		sym.syms[n] = &symbol{
			goobj:   obj,
			kind:    kind,
			id:      id,
			goname:  n,
			cgoname: "cgo_type_" + id,
			cpyname: "cpy_type_" + id,
			pyfmt:   "O&",
			pysig:   "[]" + elt.pysig,
			c2py:    "cgopy_cnv_c2py_" + id,
			py2c:    "cgopy_cnv_py2c_" + id,
		}

	case *types.Named:
		switch typ.Underlying().(type) {
		case *types.Struct:
			sym.syms[n] = &symbol{
				goobj:   obj,
				kind:    kind,
				id:      id,
				goname:  n,
				cgoname: "cgo_type_" + id,
				cpyname: "cpy_type_" + id,
				pyfmt:   "O&",
				pysig:   "object",
				c2py:    "cgopy_cnv_c2py_" + id,
				py2c:    "cgopy_cnv_py2c_" + id,
			}

		default:
			panic(fmt.Errorf("unhandled named-type: [%T]\n%#v\n", obj, t))
		}

	case *types.Pointer:
		// FIXME(sbinet): better handling?
		elm := sym.symtype(typ.Elem())
		sym.syms[n] = elm

	default:
		panic(fmt.Errorf("unhandled obj [%T]\ntype [%#v]", obj, t))
	}
}

func init() {

	type typedesc struct {
		ctype   string
		cgotype string
		pyfmt   string
		pysig   string
		c2py    string // name of converter helper C->py
		py2c    string // name of converter helper for py->c

	}

	var (
		intsize = reflect.TypeOf(int(0)).Size()

		typedescr = map[types.BasicKind]typedesc{
			types.Bool: typedesc{
				ctype:   "GoUint8",
				cgotype: "GoUint8",
				pyfmt:   "O&",
				pysig:   "bool",
				c2py:    "cgopy_cnv_c2py_bool",
				py2c:    "cgopy_cnv_py2c_bool",
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
				cgotype: "GoFloat32",
				pyfmt:   "f",
				pysig:   "float",
			},

			types.Float64: typedesc{
				ctype:   "double",
				cgotype: "GoFloat64",
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
				ctype:   "GoString",
				cgotype: "GoString",
				pyfmt:   "O&",
				pysig:   "str",
				c2py:    "cgopy_cnv_c2py_string",
				py2c:    "cgopy_cnv_py2c_string",
			},

			types.UnsafePointer: typedesc{
				ctype:   "void*",
				cgotype: "void*",
				pyfmt:   "O&",
				pysig:   "object",
			},
		}
	)

	typedescr[types.UntypedBool] = typedescr[types.Bool]
	typedescr[types.UntypedInt] = typedescr[types.Int]
	typedescr[types.UntypedRune] = typedescr[types.Rune] // FIXME(sbinet)
	typedescr[types.UntypedFloat] = typedescr[types.Float64]
	typedescr[types.UntypedComplex] = typedescr[types.Complex128]
	typedescr[types.UntypedString] = typedescr[types.String]
	typedescr[types.UntypedNil] = typedescr[types.UnsafePointer] // FIXME(sbinet)

	if intsize == 8 {
		typedescr[types.Int] = typedesc{
			ctype:   "int64_t",
			cgotype: "GoInt",
			pyfmt:   "k",
			pysig:   "int",
		}

		typedescr[types.Uint] = typedesc{
			ctype:   "uint64_t",
			cgotype: "GoUint",
			pyfmt:   "K",
			pysig:   "int",
		}
	}

	syms := make(map[string]*symbol)
	for _, n := range []string{
		"bool",
		"byte",
		"uint", "uint8", "uint16", "uint32", "uint64",
		// "uintptr", //FIXME(sbinet): how should we handle this?
		"int", "int8", "int16", "int32", "int64",
		"float32", "float64",
		"complex64", "complex128",
		"rune",
		"string",
	} {
		obj := types.Universe.Lookup(n)
		dtype, ok := typedescr[obj.Type().Underlying().(*types.Basic).Kind()]
		if !ok {
			panic(fmt.Errorf("could not lookup dtype for [%s]", n))
		}
		syms[n] = &symbol{
			goobj:   obj,
			kind:    skType,
			goname:  n,
			cgoname: dtype.cgotype,
			cpyname: dtype.ctype,
			pyfmt:   dtype.pyfmt,
			pysig:   dtype.pysig,
			c2py:    dtype.c2py,
			py2c:    dtype.py2c,
		}
	}

	{
		obj := types.Universe.Lookup("error")
		id := obj.Name()
		syms["error"] = &symbol{
			goobj:   obj,
			kind:    skType,
			goname:  "error",
			cgoname: "GoInterface",
			cpyname: "GoInterface",
			pyfmt:   "O&",
			pysig:   "object",
			c2py:    "cgopy_cnv_c2py_" + id,
			py2c:    "cgopy_cnv_py2c_" + id,
		}
	}

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
		//"nil",
	} {
		sym := *syms[o.tname]
		_, ok := typedescr[o.kind]
		if !ok {
			panic(fmt.Errorf("gopy: could not lookup dtype for [%s]", o.tname))
		}
		n := "untyped " + o.uname
		syms[n] = &sym
	}

	universe = &symtab{
		pkg:    nil,
		syms:   syms,
		parent: nil,
	}
}
