// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/doc"
	"go/token"
	"go/types"
	"reflect"
	"strings"
)

// Package ties types.Package and ast.Package together.
// Package also collects informations about structs and funcs.
type Package struct {
	pkg *types.Package
	n   int // number of entities to wrap
	sz  types.Sizes
	doc *doc.Package

	syms   *symtab
	objs   map[string]Object
	consts []Const
	vars   []Var
	types  []Type
	funcs  []Func
}

// NewPackage creates a new Package, tying types.Package and ast.Package together.
func NewPackage(pkg *types.Package, doc *doc.Package) (*Package, error) {
	universe.pkg = pkg // FIXME(sbinet)
	sz := int64(reflect.TypeOf(int(0)).Size())
	p := &Package{
		pkg:  pkg,
		n:    0,
		sz:   &types.StdSizes{sz, sz},
		doc:  doc,
		syms: newSymtab(pkg, nil),
		objs: map[string]Object{},
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

// ImportPath returns the package import path.
func (p *Package) ImportPath() string {
	return p.doc.ImportPath
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
			if o.Parent() == nil || (o.Parent() != nil && parent != "") {
				for _, typ := range p.doc.Types {
					if typ.Name != parent {
						continue
					}
					if o.Parent() == nil {
						for _, m := range typ.Methods {
							if m.Name == n {
								return m.Doc
							}
						}
					} else {
						for _, m := range typ.Funcs {
							if m.Name == n {
								return m.Doc
							}
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
				paramType := p.pysig(paramVar.Type())
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
	typs := make(map[string]Type)

	scope := p.pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if !obj.Exported() {
			continue
		}

		p.n++
		p.syms.addSymbol(obj)
	}

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
			funcs[name], err = newFuncFrom(p, "", obj, obj.Type().(*types.Signature))
			if err != nil {
				return err
			}

		case *types.TypeName:
			typs[name], err = newType(p, obj)
			if err != nil {
				return err
			}

		default:
			//TODO(sbinet)
			panic(fmt.Errorf("not yet supported: %v (%T)", obj, obj))
		}

	}

	// remove ctors from funcs.
	// add methods.
	for tname, t := range typs {
		for name, fct := range funcs {
			if fct.Return() == nil {
				continue
			}
			if fct.Return() == t.GoType() {
				delete(funcs, name)
				fct.doc = p.getDoc(tname, scope.Lookup(name))
				fct.ctor = true
				t.ctors = append(t.ctors, fct)
				typs[tname] = t
			}
		}

		ptyp := types.NewPointer(t.GoType())
		p.syms.addType(nil, ptyp)
		mset := types.NewMethodSet(ptyp)
		for i := 0; i < mset.Len(); i++ {
			meth := mset.At(i)
			if !meth.Obj().Exported() {
				continue
			}
			m, err := newFuncFrom(p, tname, meth.Obj(), meth.Type().(*types.Signature))
			if err != nil {
				return err
			}
			t.meths = append(t.meths, m)
			if isStringer(meth.Obj()) {
				t.prots |= ProtoStringer
			}
		}
		p.addType(t)
	}

	for _, fct := range funcs {
		p.addFunc(fct)
	}

	// attach docstrings to methods
	for _, n := range p.syms.names() {
		sym := p.syms.syms[n]
		if !sym.isNamed() {
			continue
		}
		switch typ := sym.GoType().(type) {
		case *types.Named:
			for i := 0; i < typ.NumMethods(); i++ {
				m := typ.Method(i)
				if !m.Exported() {
					continue
				}
				doc := p.getDoc(sym.goname, m)
				mname := types.ObjectString(m, nil)
				msym := p.syms.sym(mname)
				if msym == nil {
					panic(fmt.Errorf(
						"gopy: could not retrieve symbol for %q",
						m.FullName(),
					))
				}
				msym.doc = doc
			}
		}
	}
	return err
}

func (p *Package) addConst(obj *types.Const) {
	p.consts = append(p.consts, newConst(p, obj))
}

func (p *Package) addVar(obj *types.Var) {
	p.vars = append(p.vars, *newVarFrom(p, obj))
}

func (p *Package) addType(t Type) {
	p.types = append(p.types, t)
	p.objs[t.GoName()] = t
}

func (p *Package) addFunc(f Func) {
	p.funcs = append(p.funcs, f)
	p.objs[f.GoName()] = f
}

// Lookup returns the bind.Object corresponding to a types.Object
func (p *Package) Lookup(o types.Object) (Object, bool) {
	obj, ok := p.objs[o.Name()]
	return obj, ok
}

// pysig returns the doc-string corresponding to the given type, for pydoc.
func (p *Package) pysig(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		switch k := t.Kind(); k {
		case types.Bool:
			return "bool"
		case types.Int, types.Int8, types.Int16, types.Int32:
			return "int"
		case types.Int64:
			return "long"
		case types.Uint, types.Uint8, types.Uint16, types.Uint32:
			return "int"
		case types.Uint64:
			return "long"
		case types.Float32, types.Float64:
			return "float"
		case types.Complex64, types.Complex128:
			return "complex"
		case types.String:
			return "str"
		}
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), p.pysig(t.Elem()))
	case *types.Slice:
		return "[]" + p.pysig(t.Elem())
	case *types.Signature:
		return "callable" // FIXME(sbinet): give the exact pydoc equivalent signature ?
	case *types.Named:
		return "object" // FIXME(sbinet): give the exact python class name ?
	case *types.Map:
		return "dict" // FIXME(sbinet): give exact dict-k/v ?
	case *types.Pointer:
		return "object"
	case *types.Chan:
		return "object"
	default:
		panic(fmt.Errorf("unhandled type [%T]\n%#v\n", t, t))
	}
	panic(fmt.Errorf("unhandled type [%T]\n%#v\n", t, t))
}

// idFrom generates a new ID from a types.Object.
func idFrom(obj types.Object) string {
	return obj.Pkg().Name() + "_" + obj.Name()
}

// descFrom generates a new bind/seq description from a types.Object.
func descFrom(obj types.Object) string {
	return obj.Pkg().Path() + "." + obj.Name()
}

// Func collects informations about a go func/method.
type Func struct {
	pkg  *Package
	sig  *types.Signature
	typ  types.Type // if nil, Func was generated
	name string

	desc string // bind/seq descriptor
	id   string
	doc  string
	ret  types.Type // return type, if any
	err  bool       // true if original go func has comma-error
	ctor bool       // true if this is a newXXX function
}

func newFuncFrom(p *Package, parent string, obj types.Object, sig *types.Signature) (Func, error) {
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

	desc := obj.Pkg().Path() + "." + obj.Name()
	id := obj.Pkg().Name() + "_" + obj.Name()
	if parent != "" {
		id = obj.Pkg().Name() + "_" + parent + "_" + obj.Name()
		desc = obj.Pkg().Path() + "." + parent + "." + obj.Name()
	}

	return Func{
		pkg:  p,
		sig:  sig,
		typ:  obj.Type(),
		name: obj.Name(),
		desc: desc,
		id:   id,
		doc:  p.getDoc(parent, obj),
		ret:  ret,
		err:  haserr,
	}, nil
}

func (f Func) Package() *Package {
	return f.pkg
}

func (f Func) Descriptor() string {
	return f.desc
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

func (f Func) Signature() *types.Signature {
	return f.sig
}

func (f Func) Return() types.Type {
	return f.ret
}

func (f Func) gofmt() string {
	if f.typ != nil {
		return gofmt(f.Package().Name(), f.GoType())
	}
	return f.Descriptor() // FIXME(sbinet)
}

func (f Func) CPyName() string {
	return "cpy_func_" + f.id
}

func (f Func) CGoName() string {
	return "cgo_func_" + f.id
}

type Const struct {
	pkg *Package
	obj *types.Const
	id  string
	doc string
	f   Func
}

func newConst(p *Package, o *types.Const) Const {
	pkg := o.Pkg()
	id := pkg.Name() + "_" + o.Name()
	doc := p.getDoc("", o)

	res := types.NewVar(token.NoPos, o.Pkg(), "ret", o.Type())
	sig := types.NewSignature(nil, nil, types.NewTuple(res), false)
	fct := Func{
		pkg:  p,
		sig:  sig,
		typ:  nil,
		name: o.Name(),
		desc: p.ImportPath() + "." + o.Name() + ".get",
		id:   id + "_get",
		doc:  doc,
		ret:  o.Type(),
		err:  false,
	}

	return Const{
		pkg: p,
		obj: o,
		id:  id,
		doc: doc,
		f:   fct,
	}
}

func (c Const) Package() *Package  { return c.pkg }
func (c Const) ID() string         { return c.id }
func (c Const) Doc() string        { return c.doc }
func (c Const) GoName() string     { return c.obj.Name() }
func (c Const) GoType() types.Type { return c.obj.Type() }
func (c Const) CPyName() string    { return "cpy_const_" + c.id }
func (c Const) CGoName() string    { return "cgo_const_" + c.id }
