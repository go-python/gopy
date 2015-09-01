// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"

	"golang.org/x/tools/go/types"
)

type Var struct {
	pkg  *Package
	sym  *symbol // symbol associated with var's type
	id   string
	doc  string
	name string
}

func (v *Var) Name() string {
	return v.name
}

func newVar(p *Package, typ types.Type, objname, name, doc string) *Var {
	sym := p.syms.symtype(typ)
	if sym == nil {
		panic(fmt.Errorf("could not find symbol for type [%s]!", typ.String()))
	}
	return &Var{
		pkg:  p,
		sym:  sym,
		id:   p.Name() + "_" + objname,
		doc:  doc,
		name: name,
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

func getTypeString(t types.Type) string {
	return types.TypeString(t, func(*types.Package) string { return " " })
}

func (v *Var) GoType() types.Type {
	return v.sym.goobj.Type()
}

func (v *Var) CType() string {
	return v.sym.cpyname
}

func (v *Var) CGoType() string {
	return v.sym.cgoname
}

func (v *Var) PyCode() string {
	return v.sym.pyfmt
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
	n := v.sym.cpyname
	g.Printf("c_%[1]s = ((%[2]s*)self)->cgopy;\n", v.Name(), n)
}

func (v *Var) genRetDecl(g *printer) {
	g.Printf("%[1]s c_gopy_ret;\n", v.sym.cgoname)
}

func (v *Var) getArgParse() (string, []string) {
	addrs := make([]string, 0, 1)
	cnv := v.sym.hasConverter()
	if cnv {
		addrs = append(addrs, v.sym.py2c)
	}
	addr := "&c_" + v.Name()
	addrs = append(addrs, addr)
	return v.sym.pyfmt, addrs
}

func (v *Var) getArgBuildValue() (string, []string) {
	args := make([]string, 0, 1)
	cnv := v.sym.hasConverter()
	if cnv {
		args = append(args, ""+v.sym.c2py)
	}
	arg := "c_" + v.Name()
	if cnv {
		arg = "&" + arg
	}
	args = append(args, arg)

	return v.sym.pyfmt, args
}

func (v *Var) genFuncPreamble(g *printer) {
}

func (v *Var) getFuncArg() string {
	return "c_" + v.Name()
}
