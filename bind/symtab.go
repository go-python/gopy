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
	id      string // mangled name of entity (eg: <pkg>_<name>)
	goname  string // name of go entity
	cgoname string // name of entity for cgo
	cpyname string // name of entity for cpython

	// for types only

	pyfmt string // format string for PyArg_ParseTuple
	pybuf string // format string for 'struct'/'buffer'
	pysig string // type string for doc-signatures
	c2py  string // name of c->py converter function
	py2c  string // name of py->c converter function
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
	return s.pyfmt == "O&" && (s.c2py != "" || s.py2c != "")
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
			pybuf:   fmt.Sprintf("%d%s", typ.Len(), elt.pybuf),
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
			pybuf:   elt.pybuf,
			pysig:   "[]" + elt.pysig,
			c2py:    "cgopy_cnv_c2py_" + id,
			py2c:    "cgopy_cnv_py2c_" + id,
		}

	case *types.Named:
		switch typ := typ.Underlying().(type) {
		case *types.Struct:
			pybuf := make([]string, 0, typ.NumFields())
			for i := 0; i < typ.NumFields(); i++ {
				ftyp := typ.Field(i).Type()
				fsym := sym.symtype(ftyp)
				pybuf = append(pybuf, fsym.pybuf)
			}
			sym.syms[n] = &symbol{
				goobj:   obj,
				kind:    kind,
				id:      id,
				goname:  n,
				cgoname: "cgo_type_" + id,
				cpyname: "cpy_type_" + id,
				pyfmt:   "O&",
				pybuf:   strings.Join(pybuf, ""),
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

	look := types.Universe.Lookup
	syms := map[string]*symbol{
		"bool": {
			goobj:   look("bool"),
			kind:    skType,
			goname:  "bool",
			cgoname: "GoUint8",
			cpyname: "GoUint8",
			pyfmt:   "O&",
			pybuf:   "?",
			pysig:   "bool",
			c2py:    "cgopy_cnv_c2py_bool",
			py2c:    "cgopy_cnv_py2c_bool",
		},
		"byte": {
			goobj:   look("byte"),
			kind:    skType,
			goname:  "byte",
			cpyname: "uint8_t",
			cgoname: "GoUint8",
			pyfmt:   "b",
			pybuf:   "B",
			pysig:   "int", // FIXME(sbinet) py2/py3
		},
		"int": {
			goobj:   look("int"),
			kind:    skType,
			goname:  "int",
			cpyname: "int",
			cgoname: "GoInt",
			pyfmt:   "i",
			pybuf:   "i",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_int",
			py2c:    "cgopy_cnv_py2c_int",
		},

		"int8": {
			goobj:   look("int8"),
			kind:    skType,
			goname:  "int8",
			cpyname: "int8_t",
			cgoname: "GoInt8",
			pyfmt:   "b",
			pybuf:   "b",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_int8",
			py2c:    "cgopy_cnv_py2c_int8",
		},

		"int16": {
			goobj:   look("int16"),
			kind:    skType,
			goname:  "int16",
			cpyname: "int16_t",
			cgoname: "GoInt16",
			pyfmt:   "h",
			pybuf:   "h",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_int16",
			py2c:    "cgopy_cnv_py2c_int16",
		},

		"int32": {
			goobj:   look("int32"),
			kind:    skType,
			goname:  "int32",
			cpyname: "int32_t",
			cgoname: "GoInt32",
			pyfmt:   "i",
			pybuf:   "i",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_int32",
			py2c:    "cgopy_cnv_py2c_int32",
		},

		"int64": {
			goobj:   look("int64"),
			kind:    skType,
			goname:  "int64",
			cpyname: "int64_t",
			cgoname: "GoInt64",
			pyfmt:   "k",
			pybuf:   "q",
			pysig:   "long",
			c2py:    "cgopy_cnv_c2py_int64",
			py2c:    "cgopy_cnv_py2c_int64",
		},

		"uint": {
			goobj:   look("uint"),
			kind:    skType,
			goname:  "uint",
			cpyname: "unsigned int",
			cgoname: "GoUint",
			pyfmt:   "I",
			pybuf:   "I",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_uint",
			py2c:    "cgopy_cnv_py2c_uint",
		},

		"uint8": {
			goobj:   look("uint8"),
			kind:    skType,
			goname:  "uint8",
			cpyname: "uint8_t",
			cgoname: "GoUint8",
			pyfmt:   "B",
			pybuf:   "B",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_uint8",
			py2c:    "cgopy_cnv_py2c_uint8",
		},
		"uint16": {
			goobj:   look("uint16"),
			kind:    skType,
			goname:  "uint16",
			cpyname: "uint16_t",
			cgoname: "GoUint16",
			pyfmt:   "H",
			pybuf:   "H",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_uint16",
			py2c:    "cgopy_cnv_py2c_uint16",
		},
		"uint32": {
			goobj:   look("uint32"),
			kind:    skType,
			goname:  "uint32",
			cpyname: "uint32_t",
			cgoname: "GoUint32",
			pyfmt:   "I",
			pybuf:   "I",
			pysig:   "long",
			c2py:    "cgopy_cnv_c2py_uint32",
			py2c:    "cgopy_cnv_py2c_uint32",
		},

		"uint64": {
			goobj:   look("uint64"),
			kind:    skType,
			goname:  "uint64",
			cpyname: "uint64_t",
			cgoname: "GoUint64",
			pyfmt:   "K",
			pybuf:   "Q",
			pysig:   "long",
			c2py:    "cgopy_cnv_c2py_uint64",
			py2c:    "cgopy_cnv_py2c_uint64",
		},

		"float32": {
			goobj:   look("float32"),
			kind:    skType,
			goname:  "float32",
			cpyname: "float",
			cgoname: "GoFloat32",
			pyfmt:   "f",
			pybuf:   "f",
			pysig:   "float",
			c2py:    "cgopy_cnv_c2py_f32",
			py2c:    "cgopy_cnv_py2c_f32",
		},
		"float64": {
			goobj:   look("float64"),
			kind:    skType,
			goname:  "float64",
			cpyname: "double",
			cgoname: "GoFloat64",
			pyfmt:   "d",
			pybuf:   "d",
			pysig:   "float",
			c2py:    "cgopy_cnv_c2py_float64",
			py2c:    "cgopy_cnv_py2c_float64",
		},
		"complex64": {
			goobj:   look("complex64"),
			kind:    skType,
			goname:  "complex64",
			cpyname: "float complex",
			cgoname: "GoComplex64",
			pyfmt:   "D",
			pybuf:   "ff",
			pysig:   "complex",
			c2py:    "cgopy_cnv_c2py_complex64",
			py2c:    "cgopy_cnv_py2c_complex64",
		},
		"complex128": {
			goobj:   look("complex128"),
			kind:    skType,
			goname:  "complex128",
			cpyname: "double complex",
			cgoname: "GoComplex128",
			pyfmt:   "D",
			pybuf:   "dd",
			pysig:   "complex",
			c2py:    "cgopy_cnv_c2py_complex128",
			py2c:    "cgopy_cnv_py2c_complex128",
		},

		"string": {
			goobj:   look("string"),
			kind:    skType,
			goname:  "string",
			cpyname: "GoString",
			cgoname: "GoString",
			pyfmt:   "O&",
			pybuf:   "s",
			pysig:   "str",
			c2py:    "cgopy_cnv_c2py_string",
			py2c:    "cgopy_cnv_py2c_string",
		},

		"rune": { // FIXME(sbinet) py2/py3
			goobj:   look("rune"),
			kind:    skType,
			goname:  "rune",
			cpyname: "GoRune",
			cgoname: "GoRune",
			pyfmt:   "O&",
			pybuf:   "p",
			pysig:   "str",
			c2py:    "cgopy_cnv_c2py_rune",
			py2c:    "cgopy_cnv_py2c_rune",
		},

		"error": &symbol{
			goobj:   look("error"),
			kind:    skType,
			goname:  "error",
			cgoname: "GoInterface",
			cpyname: "GoInterface",
			pyfmt:   "O&",
			pybuf:   "PP",
			pysig:   "object",
			c2py:    "cgopy_cnv_c2py_error",
			py2c:    "cgopy_cnv_py2c_error",
		},
	}

	if reflect.TypeOf(int(0)).Size() == 8 {
		syms["int"] = &symbol{
			goobj:   look("int"),
			kind:    skType,
			goname:  "int",
			cpyname: "int64_t",
			cgoname: "GoInt",
			pyfmt:   "k",
			pybuf:   "q",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_int",
			py2c:    "cgopy_cnv_py2c_int",
		}
		syms["uint"] = &symbol{
			goobj:   look("uint"),
			kind:    skType,
			goname:  "uint",
			cpyname: "uint64_t",
			cgoname: "GoUint",
			pyfmt:   "K",
			pybuf:   "Q",
			pysig:   "int",
			c2py:    "cgopy_cnv_c2py_uint",
			py2c:    "cgopy_cnv_py2c_uint",
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
		//FIXME(sbinet): what should be the python equivalent?
		//{types.UntypedNil, "nil", "nil"},
	} {
		sym := *syms[o.tname]
		n := "untyped " + o.uname
		syms[n] = &sym
	}

	universe = &symtab{
		pkg:    nil,
		syms:   syms,
		parent: nil,
	}
}
