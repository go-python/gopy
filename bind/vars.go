// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"

	"golang.org/x/tools/go/types"
)

type Var struct {
	pkg   *Package
	id    string
	doc   string
	name  string
	typ   types.Type
	dtype typedesc
}

func (v *Var) Name() string {
	return v.name
}

func newVar(p *Package, typ types.Type, objname, name, doc string) *Var {
	return &Var{
		pkg:   p,
		id:    p.Name() + "_" + objname,
		doc:   doc,
		name:  name,
		typ:   typ,
		dtype: getTypedesc(typ),
	}
}

func newVarsFrom(p *Package, tuple *types.Tuple) []*Var {
	vars := make([]*Var, 0, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		vars = append(vars, newVarFrom(p, tuple.At(i)))
	}
	return vars
}

func newVarFrom(p *Package, v *types.Var) *Var {
	return newVar(p, v.Type(), v.Name(), v.Name(), p.getDoc("", v))
}

func getTypedesc(t types.Type) typedesc {
	switch typ := t.(type) {
	case *types.Basic:
		dtype, ok := typedescr[typ.Kind()]
		if ok {
			return dtype
		}
	case *types.Named:
		switch typ.Underlying().(type) {
		case *types.Struct:
			obj := typ.Obj()
			pkgname := obj.Pkg().Name()
			id := pkgname + "_" + obj.Name()
			return typedesc{
				ctype:   "GoPy_" + id,
				cgotype: "GoPy_" + id,
				pyfmt:   "O&",
				c2py:    "cgopy_cnv_c2py_" + id,
				py2c:    "cgopy_cnv_py2c_" + id,
			}
		case *types.Interface:
			obj := typ.Obj()
			id := obj.Name()
			if obj.Pkg() != nil {
				id = obj.Pkg().Name() + "_" + id
			}
			return typedesc{
				ctype:   "GoInterface",
				cgotype: "GoInterface",
				pyfmt:   "O&",
				c2py:    "cgopy_cnv_c2py_" + id,
				py2c:    "cgopy_cnv_py2c_" + id,
			}
		default:
			panic(fmt.Errorf("unhandled type: %#v", typ))
		}
	case *types.Pointer:
		elem := typ.Elem()
		return getTypedesc(elem)

	case *types.Signature:
		return typedesc{
			ctype:   "GoFunction",
			cgotype: "GoFunction",
			pyfmt:   "?",
		}

	case *types.Slice:
		return typedesc{
			ctype:   "GoSlice",
			cgotype: "GoSlice",
			pyfmt:   "?",
		}

	default:
		panic(fmt.Errorf("unhandled type: %#v\n", typ))
	}
	return typedesc{}
}

func (v *Var) GoType() types.Type {
	return v.typ
}

func (v *Var) CType() string {
	return v.dtype.ctype
}

func (v *Var) CGoType() string {
	return v.dtype.cgotype
}

func (v *Var) PyCode() string {
	return v.dtype.pyfmt
}

func (v *Var) isGoString() bool {
	switch typ := v.GoType().(type) {
	case *types.Basic:
		return typ.Kind() == types.String
	}
	return false
}

func (v *Var) genDecl(g *printer) {
	g.Printf("%[1]s c_%[2]s;\n", v.CGoType(), v.Name())
}

func (v *Var) genRecvDecl(g *printer) {
	g.Printf("%[1]s c_%[2]s;\n", v.CGoType(), v.Name())
}

func (v *Var) genRecvImpl(g *printer) {
	n := string(v.CGoType()[len("GoPy_"):])
	g.Printf("c_%[1]s = ((_gopy_%[2]s*)self)->cgopy;\n", v.Name(), n)
}

func (v *Var) genRetDecl(g *printer) {
	g.Printf("%[1]s c_gopy_ret;\n", v.CGoType())
}

func (v *Var) getArgParse() (string, []string) {
	addrs := make([]string, 0, 1)
	cnv := v.dtype.hasConverter()
	if cnv {
		addrs = append(addrs, v.dtype.py2c)
	}
	addr := "&c_" + v.Name()
	addrs = append(addrs, addr)
	return v.dtype.pyfmt, addrs
}

func (v *Var) getArgBuildValue() (string, []string) {
	args := make([]string, 0, 1)
	cnv := v.dtype.hasConverter()
	if cnv {
		args = append(args, ""+v.dtype.c2py)
	}
	arg := "c_" + v.Name()
	if cnv {
		arg = "&" + arg
	}
	args = append(args, arg)

	return v.dtype.pyfmt, args
}

func (v *Var) genFuncPreamble(g *printer) {
}

func (v *Var) getFuncArg() string {
	return "c_" + v.Name()
}

func (v *Var) needWrap() bool {
	typ := v.GoType()
	return needWrapType(typ)
}
