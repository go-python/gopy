// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/doc"
	"strings"

	"golang.org/x/tools/go/types"
)

// Package ties types.Package and ast.Package together.
// Package also collects informations about structs and funcs.
type Package struct {
	pkg *types.Package
	doc *doc.Package

	objs    map[string]Object
	structs []Struct
	funcs   []Func
}

// NewPackage creates a new Package, tying types.Package and ast.Package together.
func NewPackage(pkg *types.Package, doc *doc.Package) (*Package, error) {
	p := &Package{
		pkg: pkg,
		doc: doc,
	}
	err := p.process()
	if err != nil {
		return nil, err
	}
	return p, err
}

// Name returns the package name.
func (p *Package) Name() string {
	return p.pkg.Name()
}

// getDoc returns the doc string associated with types.Object
// parent is the name of the containing scope ("" for global scope)
func (p *Package) getDoc(parent string, o types.Object) string {
	n := o.Name()
	switch o.(type) {
	case *types.Const:
		for _, c := range p.doc.Consts {
			for _, cn := range c.Names {
				if n == cn {
					return c.Doc
				}
			}
		}

	case *types.Var:
		for _, v := range p.doc.Vars {
			for _, vn := range v.Names {
				if n == vn {
					return v.Doc
				}
			}
		}

	case *types.Func:
		doc := func() string {
			if o.Parent() == nil {
				for _, typ := range p.doc.Types {
					if typ.Name != parent {
						continue
					}
					for _, m := range typ.Methods {
						if m.Name == n {
							return m.Doc
						}
					}
				}
			} else {
				for _, f := range p.doc.Funcs {
					if n == f.Name {
						return f.Doc
					}
				}
			}
			return ""
		}()

		sig := o.Type().(*types.Signature)

		parseFn := func(tup *types.Tuple) []string {
			params := []string{}
			if tup == nil {
				return params
			}
			for i := 0; i < tup.Len(); i++ {
				paramVar := tup.At(i)
				paramType := pyTypeName(paramVar.Type())
				if paramVar.Name() != "" {
					paramType = fmt.Sprintf("%s %s", paramType, paramVar.Name())
				}
				params = append(params, paramType)
			}
			return params
		}

		params := parseFn(sig.Params())
		results := parseFn(sig.Results())

		paramString := strings.Join(params, ", ")
		resultString := strings.Join(results, ", ")

		//FIXME(sbinet): add receiver for methods?
		docSig := fmt.Sprintf("%s(%s) %s", o.Name(), paramString, resultString)

		if doc != "" {
			doc = fmt.Sprintf("%s\n\n%s", docSig, doc)
		} else {
			doc = docSig
		}
		return doc

	case *types.TypeName:
		for _, t := range p.doc.Types {
			if n == t.Name {
				return t.Doc
			}
		}

	default:
		// TODO(sbinet)
		panic(fmt.Errorf("not yet supported: %v (%T)", o, o))
	}

	return ""
}

// process collects informations about a go package.
func (p *Package) process() error {
	var err error

	funcs := make(map[string]Func)
	structs := make(map[string]Struct)

	scope := p.pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}

		switch obj := obj.(type) {
		case *types.Const:
			p.addConst(obj)

		case *types.Var:
			p.addVar(obj)

		case *types.Func:
			funcs[name], err = newFunc(p, "", obj, obj.Type().(*types.Signature))
			if err != nil {
				return err
			}

		case *types.TypeName:
			named := obj.Type().(*types.Named)
			switch typ := named.Underlying().(type) {
			case *types.Struct:
				structs[name], err = newStruct(p, obj)
				if err != nil {
					return err
				}

			default:
				//TODO(sbinet)
				panic(fmt.Errorf("not yet supported: %v (%T)", typ, obj))
			}

		default:
			//TODO(sbinet)
			panic(fmt.Errorf("not yet supported: %v (%T)", obj, obj))
		}

	}

	// remove ctors from funcs.
	// add methods.
	for sname, s := range structs {
		for name, fct := range funcs {
			if fct.Return() == nil {
				continue
			}
			if fct.Return() == s.GoType() {
				delete(funcs, name)
				s.ctors = append(s.ctors, fct)
				structs[sname] = s
			}
		}

		ptyp := types.NewPointer(s.GoType())
		mset := types.NewMethodSet(ptyp)
		for i := 0; i < mset.Len(); i++ {
			meth := mset.At(i)
			if !meth.Obj().Exported() {
				continue
			}
			m, err := newFunc(p, sname, meth.Obj(), meth.Type().(*types.Signature))
			if err != nil {
				return err
			}
			s.meths = append(s.meths, m)
		}
		p.structs = append(p.structs, s)
	}

	for _, fct := range funcs {
		p.funcs = append(p.funcs, fct)
	}

	return err
}

func (p *Package) addConst(obj *types.Const) {
	//TODO(sbinet)
	panic(fmt.Errorf("not yet supported: %v (%T)", obj, obj))
}

func (p *Package) addVar(obj *types.Var) {
	//TODO(sbinet)
	panic(fmt.Errorf("not yet supported: %v (%T)", obj, obj))
}

// Struct collects informations about a go struct.
type Struct struct {
	pkg *Package
	obj *types.TypeName

	id    string
	doc   string
	ctors []Func
	meths []Func
}

func newStruct(p *Package, obj *types.TypeName) (Struct, error) {
	s := Struct{
		pkg: p,
		obj: obj,
		id:  obj.Pkg().Name() + "_" + obj.Name(),
		doc: p.getDoc("", obj),
	}
	return s, nil
}

func (s Struct) Package() *Package {
	return s.pkg
}

func (s Struct) ID() string {
	return s.id
}

func (s Struct) Doc() string {
	return s.doc
}

func (s Struct) GoType() types.Type {
	return s.obj.Type()
}

func (s Struct) GoName() string {
	return s.obj.Name()
}

func (s Struct) GoObj() types.Object {
	return s.obj
}

func (s Struct) Struct() *types.Struct {
	return s.obj.Type().Underlying().(*types.Struct)
}

// Func collects informations about a go func/method.
type Func struct {
	pkg *Package
	sig *types.Signature
	obj types.Object

	id  string
	doc string
	ret types.Type // return type, if any
	err bool       // true if original go func has comma-error
}

func newFunc(p *Package, parent string, obj types.Object, sig *types.Signature) (Func, error) {
	haserr := false
	res := sig.Results()
	var ret types.Type

	switch res.Len() {
	case 2:
		if !isErrorType(res.At(1).Type()) {
			return Func{}, fmt.Errorf(
				"bind: second result value must be of type error: %s",
				obj,
			)
		}
		haserr = true
		ret = res.At(0).Type()

	case 1:
		if isErrorType(res.At(0).Type()) {
			haserr = true
			ret = nil
		} else {
			ret = res.At(0).Type()
		}
	case 0:
		ret = nil
	default:
		return Func{}, fmt.Errorf("bind: too many results to return: %v", obj)
	}

	return Func{
		pkg: p,
		sig: sig,
		obj: obj,
		id:  obj.Pkg().Name() + "_" + obj.Name(),
		doc: p.getDoc(parent, obj),
		ret: ret,
		err: haserr,
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
	return f.obj.Type()
}

func (f Func) GoName() string {
	return f.obj.Name()
}

func (f Func) GoObj() types.Object {
	return f.obj
}

func (f Func) Signature() *types.Signature {
	return f.sig
}

func (f Func) Return() types.Type {
	return f.ret
}
