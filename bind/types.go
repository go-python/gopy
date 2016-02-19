// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/token"
	"go/types"
)

type Object interface {
	Package() *Package
	ID() string
	Doc() string
	GoName() string
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

func gofmt(pkgname string, t types.Type) string {
	return types.TypeString(
		t,
		func(*types.Package) string { return pkgname },
	)
}

// Protocol encodes the various protocols a python type may implement
type Protocol int

const (
	ProtoStringer Protocol = 1 << iota
)

// Type collects informations about a go type (struct, named-type, ...)
type Type struct {
	pkg *Package
	obj *types.TypeName
	typ types.Type

	desc  string // bind/seq descriptor
	id    string
	doc   string
	ctors []Func
	meths []Func
	funcs struct {
		new  Func
		del  Func
		init Func
		str  Func
	}

	cgoname string
	cpyname string

	prots Protocol
}

func newType(p *Package, obj *types.TypeName) (Type, error) {
	sym := p.syms.symtype(obj.Type())
	if sym == nil {
		panic(fmt.Errorf("no such object [%s] in symbols table", obj.Id()))
	}

	typ := Type{
		pkg:     p,
		obj:     obj,
		typ:     obj.Type(),
		desc:    descFrom(obj),
		id:      idFrom(obj),
		doc:     p.getDoc("", obj),
		cgoname: cgonameFrom(obj.Type()),
		cpyname: cpynameFrom(obj.Type()),
	}

	recv := types.NewVar(token.NoPos, obj.Pkg(), "recv", obj.Type())

	typ.funcs.new = Func{
		pkg: p,
		sig: types.NewSignature(nil, nil, types.NewTuple(
			types.NewVar(token.NoPos, obj.Pkg(), "ret", obj.Type()),
		), false),
		typ:  nil,
		name: obj.Name(),
		desc: typ.desc + ".new",
		id:   typ.id + "_new",
		doc:  typ.doc,
		ret:  obj.Type(),
		err:  false,
	}

	styp := types.Universe.Lookup("string")
	typ.funcs.str = Func{
		pkg: p,
		sig: types.NewSignature(recv, nil, types.NewTuple(
			types.NewVar(token.NoPos, styp.Pkg(), "ret", styp.Type()),
		), false),
		typ:  nil,
		name: obj.Name(),
		desc: typ.desc + ".str",
		id:   typ.id + "_str",
		doc:  "",
		ret:  styp.Type(),
		err:  false,
	}

	return typ, nil
}

func (t Type) Package() *Package {
	return t.pkg
}

func (t Type) Descriptor() string {
	return t.desc
}

func (t Type) ID() string {
	return t.id
}

func (t Type) Doc() string {
	return t.doc
}

func (t Type) GoType() types.Type {
	return t.typ
}

func (t Type) GoName() string {
	return t.obj.Name()
}

func (t Type) CGoName() string {
	return t.cgoname
}

func (t Type) CGoTypeName() string {
	panic("not implemented")
}

func (t Type) CPyName() string {
	return t.cpyname
}

func (t Type) CPyCheck() string {
	panic("not implemented")
}

func (t Type) Struct() *types.Struct {
	s, ok := t.typ.Underlying().(*types.Struct)
	if ok {
		return s
	}
	return nil
}

func (t Type) gofmt() string {
	return gofmt(t.Package().Name(), t.GoType())
}

func (t Type) isBasic() bool {
	if _, ok := t.GoType().(*types.Basic); ok {
		return ok
	}
	return false
}

func (t Type) isArray() bool {
	if _, ok := t.GoType().(*types.Array); ok {
		return ok
	}
	if _, ok := t.GoType().Underlying().(*types.Array); ok {
		return ok
	}
	return false
}

func (t Type) isInterface() bool {
	if _, ok := t.GoType().(*types.Interface); ok {
		return ok
	}
	if _, ok := t.GoType().Underlying().(*types.Interface); ok {
		return ok
	}
	return false
}

func (t Type) isSignature() bool {
	if _, ok := t.GoType().(*types.Signature); ok {
		return ok
	}
	_, ok := t.GoType().Underlying().(*types.Signature)
	return ok
}

func (t Type) isSlice() bool {
	if _, ok := t.GoType().(*types.Slice); ok {
		return ok
	}
	if _, ok := t.GoType().Underlying().(*types.Slice); ok {
		return ok
	}
	return false
}

func (t Type) isStruct() bool {
	if _, ok := t.GoType().(*types.Struct); ok {
		return ok
	}
	if _, ok := t.GoType().Underlying().(*types.Struct); ok {
		return ok
	}
	return false
}

func (t Type) isNamed() bool {
	_, ok := t.GoType().(*types.Named)
	return ok
}

func cgonameFrom(typ types.Type) string {
	switch typ.(type) {
	case *types.Basic:
	}
	panic("unreachable")
}

func cpynameFrom(typ types.Type) string {
	switch typ.(type) {
	}
	panic("unreachable")
}

func typeFrom(t types.Type) Type {
	T, ok := Binder.types[t.String()]
	if !ok {
		panic("bind: no such type [" + t.String() + "]")
	}
	return T
}
