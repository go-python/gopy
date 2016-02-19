// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/token"
	"go/types"
)

type Var struct {
	pkg  *Package
	id   string
	doc  string
	name string
	tvar *types.Var
}

func (v *Var) Name() string {
	return v.name
}

func newVar(p *Package, typ types.Type, name, doc string) *Var {
	return &Var{
		pkg:  p,
		id:   p.Name() + "_" + name,
		doc:  doc,
		name: name,
		tvar: types.NewVar(token.NoPos, p.pkg, name, typ),
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
	return newVar(p, v.Type(), v.Name(), p.getDoc("", v))
}

func getTypeString(t types.Type) string {
	return types.TypeString(t, func(*types.Package) string { return " " })
}

func (v *Var) Package() *Package  { return v.pkg }
func (v *Var) ID() string         { return v.id }
func (v *Var) Doc() string        { return v.doc }
func (v *Var) GoName() string     { return v.name }
func (v *Var) GoType() types.Type { return v.tvar.Type() }
func (v *Var) CType() string      { return v.sym.cpyname }
func (v *Var) CGoType() string    { return v.sym.cgoname }

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

func (v *Var) needWrap() bool {
	typ := v.GoType()
	return needWrapType(typ)
}

func (v *Var) gofmt() string {
	return v.sym.gofmt()
}
