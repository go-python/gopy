// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/types"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// Package ties types.Package and ast.Package together.
// Package also collects information about specific types (structs, ifaces, etc)
type Package struct {
	pkg *types.Package
	n   int // number of entities to wrap
	sz  types.Sizes
	doc *doc.Package

	syms      *symtab // note: this is now *always* = symbols.current
	objs      map[string]Object
	consts    []*Const
	enums     []*Enum
	vars      []*Var
	structs   []*Struct
	ifaces    []*Interface
	slices    []*Slice
	maps      []*Map
	funcs     []*Func
	pyimports map[string]string // extra python imports from incidental python wrapper includes
	// calls   []*Signature // TODO: could optimize calls back into python to gen once
}

// Packages accumulates all the packages processed
var Packages []*Package

// ResetPackages resets any accumulated packages -- needed when doing tests
func ResetPackages() {
	universeMutex.Lock()
	defer universeMutex.Unlock()
	Packages = nil
	makeGoPackage()
	current = newSymtab(nil, universe)
}

// NewPackage creates a new Package, tying types.Package and ast.Package together.
func NewPackage(pkg *types.Package, doc *doc.Package) (*Package, error) {
	// protection for parallel tests
	universeMutex.Lock()
	defer universeMutex.Unlock()
	fmt.Printf("\n--- Processing package: %v ---\n", pkg.Path())
	sz := int64(reflect.TypeOf(int(0)).Size())
	p := &Package{
		pkg:       pkg,
		n:         0,
		sz:        &types.StdSizes{WordSize: sz, MaxAlign: sz},
		doc:       doc,
		syms:      current,
		objs:      map[string]Object{},
		pyimports: map[string]string{},
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

// add given path to python imports -- these packages were referenced
func (p *Package) AddPyImport(ipath string, extra bool) {
	mypath := p.pkg.Path()
	if mypath == "go" {
		return
	}
	if ipath == mypath {
		return
	}
	if p.pyimports == nil {
		p.pyimports = make(map[string]string)
	}
	if _, has := p.pyimports[ipath]; has {
		return
	}
	nm := filepath.Base(ipath)
	p.pyimports[ipath] = nm
	// if extra {
	// 	fmt.Printf("%v added py import: %v = %v\n", mypath, ipath, nm)
	// }
}

// getDoc returns the doc string associated with types.Object
// parent is the name of the containing scope ("" for global scope)
func (p *Package) getDoc(parent string, o types.Object) string {
	n := o.Name()
	switch tp := o.(type) {
	case *types.Const:
		for _, c := range p.doc.Consts {
			for _, cn := range c.Names {
				if n == cn {
					return c.Doc
				}
			}
		}

	case *types.Var:
		if tp.IsField() && parent != "" {
			// Find the package-scoped struct
			for _, typ := range p.doc.Types {
				_ = typ
				if typ.Name != parent {
					continue
				}
				// Name matches package-scoped struct.
				// Make sure it is a struct type.
				for _, spec := range typ.Decl.Specs {
					typSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					structSpec, ok := typSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					// We have the package-scoped struct matching the parent name.
					// Find the matching field.
					for _, field := range structSpec.Fields.List {
						for _, fieldName := range field.Names {
							if fieldName.Name == tp.Name() {
								return field.Doc.Text()
							}
						}
					}
				}

			}
		}
		// Otherwise just check the captured vars
		for _, v := range p.doc.Vars {
			for _, vn := range v.Names {
				if n == vn {
					return v.Doc
				}
			}
		}

	case *types.Func:
		sig := o.Type().(*types.Signature)
		_, _, _, err := isPyCompatFunc(sig)
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

		// if a function returns a type defined in the package,
		// it is organized under that type
		if doc == "" && sig.Results().Len() == 1 {
			ret := sig.Results().At(0).Type()
			if ntyp, ok := ret.(*types.Named); ok {
				tn := ntyp.Obj().Name()
				doc = func() string {
					for _, typ := range p.doc.Types {
						if typ.Name != tn {
							continue
						}
						for _, m := range typ.Funcs {
							if m.Name == n {
								return m.Doc
							}
						}
					}
					return ""
				}()
			}
		}

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
	slices := make(map[string]*Slice)
	maps := make(map[string]*Map)

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
					fmt.Println(err)
					continue
				}
				structs[name] = sv

			case *types.Basic:
				// ok. handled by p.syms-types

			case *types.Array:
				// ok. handled by p.syms-types

			case *types.Interface:
				iv, err := newInterface(p, obj)
				if err != nil {
					fmt.Println(err)
					continue
				}
				ifaces[name] = iv

			case *types.Signature:
				// ok. handled by p.syms-types

			case *types.Slice:
				sl, err := newSlice(p, obj)
				if err != nil {
					fmt.Println(err)
					continue
				}
				slices[name] = sl

			case *types.Map:
				mp, err := newMap(p, obj)
				if err != nil {
					fmt.Println(err)
					continue
				}
				maps[name] = mp

			case *types.Chan:
				continue // don't handle

			default:
				panic(fmt.Errorf("not yet supported: %v (%T)", typ, obj))
			}

		default:
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

	for sname, s := range slices {
		styp := s.GoType()
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
		p.addSlice(s)
	}

	for sname, s := range maps {
		styp := s.GoType()
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
		p.addMap(s)
	}

	for _, fct := range funcs {
		p.addFunc(fct)
	}

	return err
}

func (p *Package) findEnum(ntyp *types.Named) *Enum {
	for _, enm := range p.enums {
		if enm.typ == ntyp {
			return enm
		}
	}
	return nil
}

func (p *Package) addConst(obj *types.Const) {
	if ntyp, ok := obj.Type().(*types.Named); ok {
		enm := p.findEnum(ntyp)
		if enm != nil {
			enm.AddConst(p, obj)
			return
		} else {
			val := obj.Val().String()
			_, err := strconv.Atoi(val)
			if err == nil {
				enm, err := newEnum(p, obj)
				if err == nil {
					p.enums = append(p.enums, enm)
					return
				}
			}
		}
	}

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

func (p *Package) addSlice(slc *Slice) {
	p.slices = append(p.slices, slc)
	p.objs[slc.GoName()] = slc
}

func (p *Package) addMap(mp *Map) {
	p.maps = append(p.maps, mp)
	p.objs[mp.GoName()] = mp
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
			if ems, isstru := emss.(*Struct); isstru {
				if ems.idx > s.idx {
					nswap++
					p.structs[s.idx], p.structs[ems.idx] = p.structs[ems.idx], p.structs[s.idx]
					s.idx, ems.idx = ems.idx, s.idx
				}
			}
		}
		if nswap == 0 {
			break
		}
		// fmt.Printf("%s nswap: %v\n", p.pkg.Path(), nswap)
	}
}
