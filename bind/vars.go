// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"

	"golang.org/x/tools/go/types"
)

type Var struct {
	Var   *types.Var
	dtype typedesc
}

func newVars(tuple *types.Tuple) []*Var {
	vars := make([]*Var, 0, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		vars = append(vars, newVar(tuple.At(i)))
	}
	return vars
}

func newVar(v *types.Var) *Var {
	return &Var{
		Var:   v,
		dtype: getTypedesc(v.Type()),
	}
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
			return typedesc{
				ctype:   "GoPy_" + typ.Obj().Name(),
				cgotype: "GoPy_" + typ.Obj().Name(),
				pyfmt:   "N",
			}
		}
	case *types.Pointer:
		elem := typ.Elem()
		return getTypedesc(elem)
	default:
		panic(fmt.Errorf("unhandled type: %#v\n", typ))
	}
	return typedesc{}
}

func (v *Var) GoType() types.Type {
	return v.Var.Type()
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
	if v.isGoString() {
		g.Printf("const char* cgopy_%s;\n", v.Var.Name())
	}
	g.Printf("%[1]s c_%[2]s;\n", v.CGoType(), v.Var.Name())
}

func (v *Var) genRetDecl(g *printer) {
	if v.isGoString() {
		g.Printf("const char* cgopy_gopy_ret;\n")
	}
	g.Printf("%[1]s c_gopy_ret;\n", v.CGoType())
}

func (v *Var) getArgParse() (string, string) {
	addr := "&c_" + v.Var.Name()
	if v.isGoString() {
		addr = "&cgopy_" + v.Var.Name()
	}
	return v.dtype.pyfmt, addr
}

func (v *Var) genFuncPreamble(g *printer) {
	if v.isGoString() {
		g.Printf("c_%[1]s = CGoPy_GoString((char*)cgopy_%[1]s);\n", v.Var.Name())
	}
}

func (v *Var) getFuncArg() string {
	return "c_" + v.Var.Name()
}

func (v *Var) needWrap() bool {
	typ := v.GoType()
	return needWrapType(typ)
}
