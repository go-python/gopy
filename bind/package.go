// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/doc"
	"go/types"
	"reflect"
	"strings"
)

// Package ties types.Package and ast.Package together.
// Package also collects information about specific types (structs, ifaces, etc)
type Package struct {
	pkg *types.Package
	n   int // number of entities to wrap
	sz  types.Sizes
	doc *doc.Package

	syms    *symtab // note: this is now *always* = symbols.current
	objs    map[string]Object
	consts  []*Const
	vars    []*Var
	structs []*Struct
	ifaces  []*Interface
	funcs   []*Func
}

// accumulates all the packages processed
var Packages []*Package

// NewPackage creates a new Package, tying types.Package and ast.Package together.
func NewPackage(pkg *types.Package, doc *doc.Package) (*Package, error) {
	fmt.Printf("\n--- Processing package: %v ---\n", pkg.Path())
	sz := int64(reflect.TypeOf(int(0)).Size())
	p := &Package{
		pkg:  pkg,
		n:    0,
		sz:   &types.StdSizes{WordSize: sz, MaxAlign: sz},
		doc:  doc,
		syms: current,
		objs: map[string]Object{},
	}
	err := p.process()
	if err != nil {
		return nil, err
	}
	Packages = append(Packages, p)
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
		sig := o.Type().(*types.Signature)
		err, _, _, _ := isPyCompatFunc(sig)
		if err != nil {
			return ""
		}
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

		parseFn := func(tup *types.Tuple) []string {
			params := []string{}
			if tup == nil {
				return params
			}
			for i := 0; i < tup.Len(); i++ {
				paramVar := tup.At(i)
				paramSig := p.syms.symtype(paramVar.Type())
				if paramSig == nil {
					continue
				}
				paramType := paramSig.pysig
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

	p.syms.pkg = p.pkg
	p.syms.addImport(p.pkg)

	funcs := make(map[string]*Func)
	structs := make(map[string]*Struct)
	ifaces := make(map[string]*Interface)

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
			fv, err := newFuncFrom(p, "", obj, obj.Type().(*types.Signature))
			if err != nil {
				continue
			}
			funcs[name] = fv

		case *types.TypeName:
			named := obj.Type().(*types.Named)
			switch typ := named.Underlying().(type) {
			case *types.Struct:
				sv, err := newStruct(p, obj)
				if err != nil {
					return err
				}
				structs[name] = sv

			case *types.Basic:
				// ok. handled by p.syms-types

			case *types.Array:
				// ok. handled by p.syms-types

			case *types.Interface:
				iv, err := newInterface(p, obj)
				if err != nil {
					return err
				}
				ifaces[name] = iv

			case *types.Signature:
				// ok. handled by p.syms-types

			case *types.Slice:
				// ok. handled by p.syms-types

			case *types.Map:
				// ok. handled by p.syms-types

			case *types.Chan:
				continue // don't handle

			default:
				//TODO(sbinet)
				panic(fmt.Errorf("not yet supported: %v (%T)", typ, obj))
			}

		default:
			//TODO(sbinet)
			panic(fmt.Errorf("not yet supported: %v (%T)", obj, obj))
		}

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
					continue
				}
				msym.doc = doc
			}
		}
	}

	// remove ctors from funcs.
	// add methods.
	for sname, s := range structs {
		styp := s.GoType()
		ptyp := types.NewPointer(styp)
		p.syms.addType(nil, ptyp)
		for name, fct := range funcs {
			if !fct.Obj().Exported() {
				continue
			}
			ret := fct.Return()
			if ret == nil {
				continue
			}
			retptr, retIsPtr := ret.(*types.Pointer)

			if ret == styp || (retIsPtr && retptr.Elem() == styp) {
				delete(funcs, name)
				fct.doc = p.getDoc(sname, scope.Lookup(name))
				fct.ctor = true
				s.ctors = append(s.ctors, fct)
				structs[sname] = s
				continue
			}
		}

		ntyp, ok := styp.(*types.Named)
		if !ok {
			continue
		}

		nmeth := ntyp.NumMethods()
		for mi := 0; mi < nmeth; mi++ {
			meth := ntyp.Method(mi)
			if !meth.Exported() {
				continue
			}
			msig := meth.Type().(*types.Signature)
			m, err := newFuncFrom(p, sname, meth, msig)
			if err != nil {
				continue
			}
			s.meths = append(s.meths, m)
			if isStringer(meth) {
				s.prots |= ProtoStringer
			}
		}
		p.addStruct(s)
	}

	for iname, ifc := range ifaces {
		mset := types.NewMethodSet(ifc.GoType())
		for i := 0; i < mset.Len(); i++ {
			meth := mset.At(i)
			if !meth.Obj().Exported() {
				continue
			}
			m, err := newFuncFrom(p, iname, meth.Obj(), meth.Type().(*types.Signature))
			if err != nil {
				continue
			}
			ifc.meths = append(ifc.meths, m)
		}
		p.addInterface(ifc)
	}

	for _, fct := range funcs {
		p.addFunc(fct)
	}

	return err
}

func (p *Package) addConst(obj *types.Const) {
	nc, err := newConst(p, obj)
	if err == nil {
		p.consts = append(p.consts, nc)
	}
}

func (p *Package) addVar(obj *types.Var) {
	nv, err := newVarFrom(p, obj)
	if err == nil {
		p.vars = append(p.vars, nv)
	}
}

func (p *Package) addStruct(s *Struct) {
	p.structs = append(p.structs, s)
	s.idx = len(p.structs) - 1
	p.objs[s.GoName()] = s
}

func (p *Package) addInterface(ifc *Interface) {
	p.ifaces = append(p.ifaces, ifc)
	p.objs[ifc.GoName()] = ifc
}

func (p *Package) addFunc(f *Func) {
	p.funcs = append(p.funcs, f)
	p.objs[f.GoName()] = f
}

// Lookup returns the bind.Object corresponding to a types.Object
func (p *Package) Lookup(o types.Object) (Object, bool) {
	obj, ok := p.objs[o.Name()]
	return obj, ok
}

func (p *Package) sortStructEmbeds() {
	for {
		nswap := 0
		for _, s := range p.structs {
			emb := s.FirstEmbed()
			if emb == nil {
				continue
			}
			emss, ok := p.objs[emb.goname]
			if !ok {
				continue
			}
			ems := emss.(*Struct)
			if ems.idx > s.idx {
				nswap++
				p.structs[s.idx], p.structs[ems.idx] = p.structs[ems.idx], p.structs[s.idx]
				s.idx, ems.idx = ems.idx, s.idx
			}
		}
		if nswap == 0 {
			break
		}
		// fmt.Printf("%s nswap: %v\n", p.pkg.Path(), nswap)
	}
}
