// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
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
	switch typ := typ.(type) {
	case *types.Basic:
		return false
	case *types.Struct:
		return true
	case *types.Named:
		switch ut := typ.Underlying().(type) {
		case *types.Basic:
			return false
		default:
			return needWrapType(ut)
		}
	case *types.Array:
		return true
	case *types.Map:
		return true
	case *types.Slice:
		return true
	case *types.Interface:
		wrap := true
		if typ.Underlying() == universe.syms["error"].GoType().Underlying() {
			wrap = false
		}
		return wrap
	case *types.Signature:
		return true
	case *types.Pointer:
		return needWrapType(typ.Elem())
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////////
//  Struct

// Protocol encodes the various protocols a python type may implement
type Protocol int

const (
	ProtoStringer Protocol = 1 << iota
)

// Struct collects informations about a go struct.
type Struct struct {
	pkg *Package
	sym *symbol
	obj *types.TypeName

	id    string
	doc   string
	ctors []Func
	meths []Func

	prots Protocol
}

func newStruct(p *Package, obj *types.TypeName) (Struct, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		panic(fmt.Errorf("no such object [%s] in symbols table", obj.Id()))
	}
	sym.doc = p.getDoc("", obj)
	s := Struct{
		pkg: p,
		sym: sym,
		obj: obj,
	}
	return s, nil
}

func (s Struct) Package() *Package {
	return s.pkg
}

func (s Struct) ID() string {
	return s.sym.id
}

func (s Struct) Doc() string {
	return s.sym.doc
}

func (s Struct) GoType() types.Type {
	return s.sym.GoType()
}

func (s Struct) GoName() string {
	return s.sym.goname
}

func (s Struct) Struct() *types.Struct {
	return s.sym.GoType().Underlying().(*types.Struct)
}

///////////////////////////////////////////////////////////////////////////////////
//  Interface

// Interface collects informations about a go struct.
type Interface struct {
	pkg *Package
	sym *symbol
	obj *types.TypeName

	id    string
	doc   string
	meths []Func
}

func newInterface(p *Package, obj *types.TypeName) (Interface, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		panic(fmt.Errorf("no such object [%s] in symbols table", obj.Id()))
	}
	sym.doc = p.getDoc("", obj)
	s := Interface{
		pkg: p,
		sym: sym,
		obj: obj,
	}
	return s, nil
}

func (s Interface) Package() *Package {
	return s.pkg
}

func (s Interface) ID() string {
	return s.sym.id
}

func (s Interface) Doc() string {
	return s.sym.doc
}

func (s Interface) GoType() types.Type {
	return s.sym.GoType()
}

func (s Interface) GoName() string {
	return s.sym.goname
}

func (s Interface) Interface() *types.Interface {
	return s.sym.GoType().Underlying().(*types.Interface)
}

///////////////////////////////////////////////////////////////////////////////////
//  Signature

// A Signature represents a (non-builtin) function or method type.
type Signature struct {
	ret  []*Var
	args []*Var
	recv *Var
}

func newSignatureFrom(pkg *Package, sig *types.Signature) *Signature {
	var recv *Var
	if sig.Recv() != nil {
		recv = newVarFrom(pkg, sig.Recv())
	}

	return &Signature{
		ret:  newVarsFrom(pkg, sig.Results()),
		args: newVarsFrom(pkg, sig.Params()),
		recv: recv,
	}
}

func newSignature(pkg *Package, recv *Var, params, results []*Var) *Signature {
	return &Signature{
		ret:  results,
		args: params,
		recv: recv,
	}
}

func (sig *Signature) Results() []*Var {
	return sig.ret
}

func (sig *Signature) Params() []*Var {
	return sig.args
}

func (sig *Signature) Recv() *Var {
	return sig.recv
}

///////////////////////////////////////////////////////////////////////////////////
//  Func

// Func collects informations about a go func/method.
type Func struct {
	pkg  *Package
	sig  *Signature
	typ  types.Type
	name string

	id   string
	doc  string
	ret  types.Type // return type, if any
	err  bool       // true if original go func has comma-error
	ctor bool       // true if this is a newXXX function
}

func newFuncFrom(p *Package, parent string, obj types.Object, sig *types.Signature) (Func, error) {
	err, ret, haserr := isPyCompatFunc(sig)
	if err != nil {
		return Func{}, err
	}

	id := obj.Pkg().Name() + "_" + obj.Name()
	if parent != "" {
		id = obj.Pkg().Name() + "_" + parent + "_" + obj.Name()
	}

	return Func{
		pkg:  p,
		sig:  newSignatureFrom(p, sig),
		typ:  obj.Type(),
		name: obj.Name(),
		id:   id,
		doc:  p.getDoc(parent, obj),
		ret:  ret,
		err:  haserr,
	}, nil
}

func (f Func) Package() *Package {
	return f.pkg
}

func (f Func) ID() string {
	return f.id
}

func (f Func) Doc() string {
	return f.doc
}

func (f Func) GoType() types.Type {
	return f.typ
}

func (f Func) GoName() string {
	return f.name
}

func (f Func) GoFmt() string {
	return f.pkg.Name() + "." + f.name
}

func (f Func) Signature() *Signature {
	return f.sig
}

func (f Func) Return() types.Type {
	return f.ret
}

///////////////////////////////////////////////////////////////////////////////////
//  Const

type Const struct {
	pkg *Package
	sym *symbol
	obj *types.Const
	id  string
	doc string
	f   Func
}

func newConst(p *Package, o *types.Const) Const {
	pkg := o.Pkg()
	sym := p.syms.symtype(o.Type())
	id := pkg.Name() + "_" + o.Name()
	doc := p.getDoc("", o)

	res := []*Var{newVar(p, o.Type(), "ret", o.Name(), doc)}
	sig := newSignature(p, nil, nil, res)
	fct := Func{
		pkg:  p,
		sig:  sig,
		typ:  nil,
		name: o.Name(),
		id:   id + "_get",
		doc:  doc,
		ret:  o.Type(),
		err:  false,
	}

	return Const{
		pkg: p,
		sym: sym,
		obj: o,
		id:  id,
		doc: doc,
		f:   fct,
	}
}

func (c Const) ID() string         { return c.id }
func (c Const) Doc() string        { return c.doc }
func (c Const) GoName() string     { return c.obj.Name() }
func (c Const) GoType() types.Type { return c.obj.Type() }

///////////////////////////////////////////////////////////////////////////////////
//  Var

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
		typname, _ := p.syms.typeNamePkg(typ)
		typobj := p.syms.pkg.Scope().Lookup(typname)
		if typobj != nil {
			p.syms.addSymbol(typobj)
		} else {
			if ntyp, ok := typ.(*types.Named); ok {
				p.syms.addType(ntyp.Obj(), typ)
			} else {
				p.syms.addType(nil, typ)
			}
		}
		sym = p.syms.symtype(typ)
		if sym == nil {
			panic(fmt.Errorf("could not find symbol for type [%s]!", typ.String()))
		}
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
	return v.sym.gotyp
}

func (v *Var) CType() string {
	return v.sym.cpyname
}

func (v *Var) CGoType() string {
	return v.sym.cgoname
}

func (v *Var) isGoString() bool {
	switch typ := v.GoType().(type) {
	case *types.Basic:
		return typ.Kind() == types.String
	}
	return false
}
