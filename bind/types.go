// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"sort"
	"strconv"
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

// Struct collects information about a go struct.
type Struct struct {
	pkg *Package
	sym *symbol
	obj *types.TypeName

	id    string
	doc   string
	ctors []*Func
	meths []*Func
	idx   int // index position in list of structs

	prots Protocol
}

func newStruct(p *Package, obj *types.TypeName) (*Struct, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		return nil, fmt.Errorf("no such object [%s] in symbols table", obj.Id())
	}
	sym.doc = p.getDoc("", obj)
	s := &Struct{
		pkg: p,
		sym: sym,
		obj: obj,
	}
	return s, nil
}

func (s *Struct) Obj() *types.TypeName {
	return s.obj
}

func (s *Struct) Package() *Package {
	return s.pkg
}

func (s *Struct) ID() string {
	return s.sym.id
}

func (s *Struct) Doc() string {
	return s.sym.doc
}

func (s *Struct) GoType() types.Type {
	return s.sym.GoType()
}

func (s *Struct) GoName() string {
	return s.sym.goname
}

func (s *Struct) Struct() *types.Struct {
	return s.sym.GoType().Underlying().(*types.Struct)
}

// FirstEmbed returns the first field if it is embedded,
// supporting convention of placing embedded "parent" types first
func (s *Struct) FirstEmbed() *symbol {
	st := s.Struct()
	numFields := st.NumFields()
	if numFields == 0 {
		return nil
	}
	f := s.Struct().Field(0)
	if !f.Embedded() {
		return nil
	}
	ftyp := current.symtype(f.Type())
	if ftyp == nil {
		return nil
	}
	return ftyp
}

///////////////////////////////////////////////////////////////////////////////////
//  Interface

// Interface collects information about a go interface.
type Interface struct {
	pkg *Package
	sym *symbol
	obj *types.TypeName

	id    string
	doc   string
	meths []*Func
}

func newInterface(p *Package, obj *types.TypeName) (*Interface, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		return nil, fmt.Errorf("no such object [%s] in symbols table", obj.Id())
	}
	sym.doc = p.getDoc("", obj)
	s := &Interface{
		pkg: p,
		sym: sym,
		obj: obj,
	}
	return s, nil
}

func (it *Interface) Package() *Package {
	return it.pkg
}

func (it *Interface) ID() string {
	return it.sym.id
}

func (it *Interface) Doc() string {
	return it.sym.doc
}

func (it *Interface) GoType() types.Type {
	return it.sym.GoType()
}

func (it *Interface) GoName() string {
	return it.sym.goname
}

func (it *Interface) Interface() *types.Interface {
	return it.sym.GoType().Underlying().(*types.Interface)
}

///////////////////////////////////////////////////////////////////////////////////
//  Slice

// Slice collects information about a go slice.
type Slice struct {
	pkg *Package
	sym *symbol
	obj *types.TypeName

	id    string
	doc   string
	meths []*Func
	prots Protocol
}

func newSlice(p *Package, obj *types.TypeName) (*Slice, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		return nil, fmt.Errorf("no such object [%s] in symbols table", obj.Id())
	}
	sym.doc = p.getDoc("", obj)
	s := &Slice{
		pkg: p,
		sym: sym,
		obj: obj,
	}
	return s, nil
}

func (it *Slice) Package() *Package {
	return it.pkg
}

func (it *Slice) ID() string {
	return it.sym.id
}

func (it *Slice) Doc() string {
	return it.sym.doc
}

func (it *Slice) GoType() types.Type {
	return it.sym.GoType()
}

func (it *Slice) GoName() string {
	return it.sym.goname
}

func (it *Slice) Slice() *types.Slice {
	return it.sym.GoType().Underlying().(*types.Slice)
}

///////////////////////////////////////////////////////////////////////////////////
//  Map

// Map collects information about a go map.
type Map struct {
	pkg *Package
	sym *symbol
	obj *types.TypeName

	id    string
	doc   string
	meths []*Func
	prots Protocol
}

func newMap(p *Package, obj *types.TypeName) (*Map, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		return nil, fmt.Errorf("no such object [%s] in symbols table", obj.Id())
	}
	sym.doc = p.getDoc("", obj)
	s := &Map{
		pkg: p,
		sym: sym,
		obj: obj,
	}
	return s, nil
}

func (it *Map) Package() *Package {
	return it.pkg
}

func (it *Map) ID() string {
	return it.sym.id
}

func (it *Map) Doc() string {
	return it.sym.doc
}

func (it *Map) GoType() types.Type {
	return it.sym.GoType()
}

func (it *Map) GoName() string {
	return it.sym.goname
}

func (it *Map) Map() *types.Map {
	return it.sym.GoType().Underlying().(*types.Map)
}

///////////////////////////////////////////////////////////////////////////////////
//  Signature

// A Signature represents a (non-builtin) function or method type.
type Signature struct {
	ret  []*Var
	args []*Var
	recv *Var
}

func newSignatureFrom(pkg *Package, sig *types.Signature) (*Signature, error) {
	var recv *Var
	var err error
	if sig.Recv() != nil {
		recv, err = newVarFrom(pkg, sig.Recv())
		if err != nil {
			return nil, err
		}
	}

	rv, err := newVarsFrom(pkg, sig.Results())
	if err != nil {
		return nil, err
	}
	av, err := newVarsFrom(pkg, sig.Params())
	if err != nil {
		return nil, err
	}

	return &Signature{
		ret:  rv,
		args: av,
		recv: recv,
	}, nil
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

// Func collects information about a go func/method.
type Func struct {
	pkg  *Package
	sig  *Signature
	typ  types.Type
	obj  types.Object
	name string

	id     string
	doc    string
	ret    types.Type // return type, if any
	err    bool       // true if original go func has comma-error
	ctor   bool       // true if this is a newXXX function
	hasfun bool       // true if this function has a function argument
}

func newFuncFrom(p *Package, parent string, obj types.Object, sig *types.Signature) (*Func, error) {
	ret, haserr, hasfun, err := isPyCompatFunc(sig)
	if err != nil {
		return nil, err
	}

	id := obj.Pkg().Name() + "_" + obj.Name()
	if parent != "" {
		id = obj.Pkg().Name() + "_" + parent + "_" + obj.Name()
	}

	sv, err := newSignatureFrom(p, sig)
	if err != nil {
		return nil, err
	}

	return &Func{
		obj:    obj,
		pkg:    p,
		sig:    sv,
		typ:    obj.Type(),
		name:   obj.Name(),
		id:     id,
		doc:    p.getDoc(parent, obj),
		ret:    ret,
		err:    haserr,
		hasfun: hasfun,
	}, nil

	// TODO: could optimize by generating code once for each type of callback
	// but probably not worth the effort required to link everything up..
	// if hasfun {
	// 	args := sv.args
	// 	for i := range args {
	// 		arg := args[i]
	// 		if arg.sym.isSignature() {
	//        // TODO: need to make sure not already on the list, etc
	// 			p.calls = append(p.calls, arg)
	// 		}
	// 	}
	// }
} //

func (f *Func) Obj() types.Object {
	return f.obj
}

func (f *Func) Package() *Package {
	return f.pkg
}

func (f *Func) ID() string {
	return f.id
}

func (f *Func) Doc() string {
	return f.doc
}

func (f *Func) GoType() types.Type {
	return f.typ
}

func (f *Func) GoName() string {
	return f.name
}

func (f *Func) GoFmt() string {
	return f.pkg.Name() + "." + f.name
}

func (f *Func) Signature() *Signature {
	return f.sig
}

func (f *Func) Return() types.Type {
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
	val string
}

func newConst(p *Package, o *types.Const) (*Const, error) {
	pkg := o.Pkg()
	sym := p.syms.symtype(o.Type())
	id := pkg.Name() + "_" + o.Name()
	doc := p.getDoc("", o)
	val := o.Val().String()

	return &Const{
		pkg: p,
		sym: sym,
		obj: o,
		id:  id,
		doc: doc,
		val: val,
	}, nil
}

func (c *Const) ID() string         { return c.id }
func (c *Const) Doc() string        { return c.doc }
func (c *Const) GoName() string     { return c.obj.Name() }
func (c *Const) GoType() types.Type { return c.obj.Type() }

///////////////////////////////////////////////////////////////////////////////////
//  Enum

type Enum struct {
	pkg   *Package
	sym   *symbol
	obj   *types.Const // first one -- random..
	typ   *types.Named
	id    string
	doc   string
	items []*Const
}

func newEnum(p *Package, o *types.Const) (*Enum, error) {
	pkg := o.Pkg()
	sym := p.syms.symtype(o.Type())
	id := pkg.Name() + "_" + o.Name()
	typ := o.Type().(*types.Named)
	doc := p.getDoc("", typ.Obj())

	e := &Enum{
		pkg: p,
		sym: sym,
		obj: o,
		typ: typ,
		id:  id,
		doc: doc,
	}
	e.AddConst(p, o)
	return e, nil
}

func (e *Enum) ID() string         { return e.id }
func (e *Enum) Doc() string        { return e.doc }
func (e *Enum) GoName() string     { return e.obj.Name() }
func (e *Enum) GoType() types.Type { return e.obj.Type() }

func (e *Enum) AddConst(p *Package, o *types.Const) (*Const, error) {
	c, err := newConst(p, o)
	if err == nil {
		e.items = append(e.items, c)
	}
	return c, err
}

func (e *Enum) SortConsts() {
	sort.Slice(e.items, func(i, j int) bool {
		iv, _ := strconv.Atoi(e.items[i].val)
		jv, _ := strconv.Atoi(e.items[j].val)
		return iv < jv
	})
}

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

func newVar(p *Package, typ types.Type, objname, name, doc string) (*Var, error) {
	sym := p.syms.symtype(typ)
	if sym == nil {
		typname, _, _ := p.syms.typeNamePkg(typ)
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
			return nil, fmt.Errorf("could not find symbol for type: %s!", typ.String())
		}
	}
	return &Var{
		pkg:  p,
		sym:  sym,
		id:   p.Name() + "_" + objname,
		doc:  doc,
		name: name,
	}, nil
}

func newVarsFrom(p *Package, tuple *types.Tuple) ([]*Var, error) {
	vars := make([]*Var, 0, tuple.Len())
	var lsterr error
	for i := 0; i < tuple.Len(); i++ {
		nv, err := newVarFrom(p, tuple.At(i))
		if err != nil {
			lsterr = err
		} else {
			vars = append(vars, nv)
		}
	}
	return vars, lsterr
}

func newVarFrom(p *Package, v *types.Var) (*Var, error) {
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
