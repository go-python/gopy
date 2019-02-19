// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"hash/fnv"
	"reflect"
	"sort"
	"strings"
	"sync"
)

var (
	universeMutex sync.Mutex
	universe      *symtab
)

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("0x%d", h.Sum32())
}

// symkind describes the kinds of symbol
type symkind int

const (
	skConst symkind = 1 << iota
	skVar
	skFunc
	skType
	skArray
	skBasic
	skInterface
	skMap
	skNamed
	skPointer
	skSignature
	skSlice
	skStruct
	skString
)

var (
	symkinds = map[string]symkind{
		"const":     skConst,
		"var":       skVar,
		"func":      skFunc,
		"type":      skType,
		"array":     skArray,
		"basic":     skBasic,
		"interface": skInterface,
		"map":       skMap,
		"named":     skNamed,
		"pointer":   skPointer,
		"signature": skSignature,
		"slice":     skSlice,
		"struct":    skStruct,
		"string":    skString,
	}
)

func (k symkind) String() string {
	str := []string{}
	for n, v := range symkinds {
		if (k & v) != 0 {
			str = append(str, n)
		}
	}
	sort.Strings(str)
	return strings.Join(str, "|")
}

// symbol is an exported symbol in a go package
type symbol struct {
	kind         symkind
	gopkg        *types.Package
	goobj        types.Object
	gotyp        types.Type
	doc          string
	id           string // mangled name of entity (eg: <pkg>_<name>)
	goname       string // name of go entity
	pyname       string // name of python entity (nonptrname, cleaned-up slice name)
	cgoname      string // type name of entity for cgo -- handle for objs
	cpyname      string // type name of entity for cgo - python -- handle..
	pysig        string // type string for doc-signatures
	go2py        string // name of go->py converter function
	py2go        string // name of py->go converter function
	py2goParenEx string // extra parentheses needed at end of py2go (beyond 1 default)
	zval         string // zero value representation
}

func isPrivate(s string) bool {
	return (strings.ToLower(s[0:1]) == s[0:1])
}

func (s *symbol) isType() bool {
	return (s.kind & skType) != 0
}

func (s *symbol) isNamed() bool {
	return (s.kind & skNamed) != 0
}

func (s *symbol) isBasic() bool {
	return (s.kind & skBasic) != 0
}

func (s *symbol) isNamedBasic() bool {
	if !s.isNamed() {
		return false
	}
	_, ok := s.gotyp.Underlying().(*types.Basic)
	if ok {
		return true
	}
	return false
}

func (s *symbol) isArray() bool {
	return (s.kind & skArray) != 0
}

func (s *symbol) isInterface() bool {
	return (s.kind & skInterface) != 0
}

func (s *symbol) isSignature() bool {
	return (s.kind & skSignature) != 0
}

func (s *symbol) isMap() bool {
	return (s.kind & skMap) != 0
}

func (s *symbol) isPySequence() bool {
	return s.isArray() || s.isSlice() || s.isMap()
}

func (s *symbol) isSlice() bool {
	return (s.kind & skSlice) != 0
}

func (s *symbol) isStruct() bool {
	return (s.kind & skStruct) != 0
}

func (s *symbol) isPointer() bool {
	return (s.kind & skPointer) != 0
}

func (s *symbol) isPtrOrIface() bool {
	return s.isPointer() || s.isInterface()
}

func (s *symbol) hasHandle() bool {
	return !s.isBasic()
}

func (s *symbol) hasConverter() bool {
	return (s.go2py != "" || s.py2go != "")
}

func (s *symbol) pkgname() string {
	if s.gopkg == nil {
		return ""
	}
	return s.gopkg.Name()
}

func (s *symbol) GoType() types.Type {
	if s.goobj != nil {
		return s.goobj.Type()
	}
	return s.gotyp
}

func (s *symbol) cgotypename() string {
	typ := s.gotyp
	switch typ := typ.(type) {
	case *types.Basic:
		n := typ.Name()
		if strings.HasPrefix(n, "untyped ") {
			n = string(n[len("untyped "):])
		}
		return n
	case *types.Named:
		obj := s.goobj
		switch typ.Underlying().(type) {
		case *types.Struct:
			return s.cgoname
		case *types.Interface:
			if obj.Name() == "error" {
				return "error"
			}
		}
	}
	return s.cgoname
}

func (s *symbol) gofmt() string {
	return types.TypeString(
		s.GoType(),
		func(pkg *types.Package) string { return pkg.Name() })
}

// symtab is a table of symbols in a go package
type symtab struct {
	pkg     *types.Package
	syms    map[string]*symbol
	imports map[string]string // other packages to import b/c we refer to their types
	parent  *symtab
}

func newSymtab(pkg *types.Package, parent *symtab) *symtab {
	if parent == nil {
		parent = universe
	}
	s := &symtab{
		pkg:     pkg,
		syms:    make(map[string]*symbol),
		imports: make(map[string]string),
		parent:  parent,
	}
	return s
}

func (sym *symtab) names() []string {
	names := make([]string, 0, len(sym.syms))
	for n := range sym.syms {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (sym *symtab) sym(n string) *symbol {
	s, ok := sym.syms[n]
	if ok {
		return s
	}
	if sym.parent != nil {
		return sym.parent.sym(n)
	}
	return nil
}

func (sym *symtab) typeof(n string) *symbol {
	s := sym.sym(n)
	switch s.kind {
	case skVar, skConst:
		tname := sym.typename(s.goobj.Type(), nil)
		return sym.sym(tname)
	case skFunc:
		//FIXME(sbinet): really?
		return s
	case skType:
		return s
	default:
		panic(fmt.Errorf("unhandled symbol kind (%v)", s.kind))
	}
}

func (sym *symtab) typename(t types.Type, pkg *types.Package) string {
	if pkg == nil {
		return types.TypeString(t, nil)
	}
	return types.TypeString(t, types.RelativeTo(pkg))
}

func (sym *symtab) symtype(t types.Type) *symbol {
	tname := sym.typename(t, nil)
	return sym.sym(tname)
}

func (sym *symtab) addSymbol(obj types.Object) error {
	fn := types.ObjectString(obj, nil)
	n := obj.Name()
	pkg := obj.Pkg()
	id := n
	pyname := n
	if pkg != nil {
		id = pkg.Name() + "_" + n
	}
	switch obj.(type) {
	case *types.Const:
		sym.syms[fn] = &symbol{
			gopkg:   pkg,
			goobj:   obj,
			kind:    skConst,
			id:      id,
			goname:  n,
			pyname:  pyname,
			cgoname: n, // todo: type names!
			cpyname: n,
		}
		return sym.addType(obj, obj.Type())

	case *types.Var:
		sym.syms[fn] = &symbol{
			gopkg:   pkg,
			goobj:   obj,
			kind:    skVar,
			id:      id,
			goname:  n,
			pyname:  pyname,
			cgoname: n,
			cpyname: n,
		}
		return sym.addType(obj, obj.Type())

	case *types.Func:
		sig := obj.Type().Underlying().(*types.Signature)
		err, _, _ := isPyCompatFunc(sig)
		if err == nil {
			sym.syms[fn] = &symbol{
				gopkg:   pkg,
				goobj:   obj,
				kind:    skFunc,
				id:      id,
				goname:  n,
				pyname:  pyname,
				cgoname: n,
				cpyname: n,
			}
			err := sym.processTuple(sig.Params())
			if err != nil {
				return err
			}
			return sym.processTuple(sig.Results())
		}
	case *types.TypeName:
		return sym.addType(obj, obj.Type())

	default:
		return fmt.Errorf("gopy: handled object [%#v]", obj)
	}
	return nil
}

func (sym *symtab) processTuple(tuple *types.Tuple) error {
	if tuple == nil {
		return nil
	}
	for i := 0; i < tuple.Len(); i++ {
		ivar := tuple.At(i)
		ityp := ivar.Type()
		isym := sym.symtype(ityp)
		if isym == nil {
			err := sym.addType(ivar, ityp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// isPyCompatFunc checks if function signature is a python-compatible function.
// Returns nil if function is compatible, err message if not.
// Also returns the return type of the function
// extra bool is true if 2nd arg is an error type, which is only
// supported form of multi-return-value functions
func isPyCompatFunc(sig *types.Signature) (error, types.Type, bool) {
	haserr := false
	res := sig.Results()
	var ret types.Type

	switch res.Len() {
	case 2:
		if !isErrorType(res.At(1).Type()) {
			return fmt.Errorf("gopy: second result value must be of type error: %s", sig.String()), ret, haserr
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
		return fmt.Errorf("gopy: too many results to return: %s", sig.String()), ret, haserr
	}

	if _, isChan := ret.(*types.Chan); isChan {
		return fmt.Errorf("gopy: channel types not supported: %s", sig.String()), ret, haserr
	}

	args := sig.Params()
	nargs := args.Len()
	for i := 0; i < nargs; i++ {
		arg := args.At(i)
		argt := arg.Type()
		if _, isSig := argt.(*types.Signature); isSig {
			return fmt.Errorf("gopy: func args (signature) not supported: %s", sig.String()), ret, haserr
		}
		if _, isChan := ret.(*types.Chan); isChan {
			return fmt.Errorf("gopy: channel types not supported: %s", sig.String()), ret, haserr
		}
	}
	return nil, ret, haserr
}

// goToPyName translates a go type name to a safe python-usable
// name -- converts [] -> SliceOf_, [*] -> MapOf_*_
func goToPyName(gon string) string {
	pyn := strings.Replace(gon, "[]", "SliceOf_", -1)
	pyn = strings.Replace(pyn, "[", "MapOf_", -1)
	pyn = strings.Replace(pyn, "]", "_", -1)
	pyn = strings.Replace(pyn, "{}", "_", -1)
	return pyn
}

// typePkg gets the package for a given types.Type
func typePkg(t types.Type) *types.Package {
	if tn, ok := t.(*types.Named); ok {
		return tn.Obj().Pkg()
	}
	switch tt := t.(type) {
	case *types.Pointer:
		return typePkg(tt.Elem())
	}
	return nil
}

// typeNamePkg gets the name and package for a given types.Type -- deals with
// naming for types in other packages, and adds those packages to imports paths.
// Falls back on sym.pkg if no other package info avail.
func (sym *symtab) typeNamePkg(t types.Type) (string, *types.Package) {
	n := sym.typename(t, sym.pkg)
	var pkg *types.Package
	if lidx := strings.LastIndex(n, "/"); lidx > 0 {
		qnm := n[lidx+1:]
		n = strings.Replace(qnm, ".", "_", 1)
		pkg = typePkg(t)
		if pkg != nil {
			ip := pkg.Path()
			sym.imports[ip] = ip
		}
	} else if pidx := strings.LastIndex(n, "."); pidx > 0 {
		n = strings.Replace(n, ".", "_", 1)
		pkg = typePkg(t)
		if pkg != nil {
			ip := pkg.Path()
			sym.imports[ip] = ip
		}
	}
	if pkg == nil {
		pkg = sym.pkg
	}
	return n, pkg
}

func (sym *symtab) addType(obj types.Object, t types.Type) error {
	fn := sym.typename(t, nil)
	n, pkg := sym.typeNamePkg(t)
	id := n
	if pkg != nil {
		id = pkg.Name() + "_" + n
	}
	kind := skType
	switch typ := t.(type) {
	case *types.Basic:
		kind |= skBasic
		styp := sym.symtype(typ)
		if styp == nil {
			return fmt.Errorf("builtin type not already known [%s]!", n)
		}

	case *types.Pointer:
		return sym.addPointerType(pkg, obj, t, kind, id, n)

	case *types.Array:
		return sym.addArrayType(pkg, obj, t, kind, id, n)

	case *types.Slice:
		return sym.addSliceType(pkg, obj, t, kind, id, n)

	case *types.Map:
		return sym.addMapType(pkg, obj, t, kind, id, n)

	case *types.Signature:
		return sym.addSignatureType(pkg, obj, t, kind, id, n)

	case *types.Interface:
		return sym.addInterfaceType(pkg, obj, t, kind, id, n)

	case *types.Chan:
		return fmt.Errorf("gopy: channel type not supported: %s\n", n)

	case *types.Named:
		kind |= skNamed
		var err error
		switch st := typ.Underlying().(type) {
		case *types.Struct:
			err = sym.addStructType(pkg, obj, t, kind, id, n)

		case *types.Basic:
			styp := sym.symtype(st)
			py2go := pkg.Name() + "." + sym.typename(t, pkg)
			py2goParEx := ""
			if styp.py2go != "" {
				py2go += "(" + styp.py2go
				py2goParEx = ")"
			}

			sym.syms[fn] = &symbol{
				gopkg:        pkg,
				goobj:        obj,
				gotyp:        t,
				kind:         kind | skBasic,
				id:           id,
				goname:       styp.goname,
				pyname:       styp.pyname,
				cgoname:      styp.cgoname,
				cpyname:      styp.cpyname,
				pysig:        styp.pysig,
				go2py:        styp.cgoname,
				py2go:        py2go,
				py2goParenEx: py2goParEx,
				zval:         styp.zval,
			}

		case *types.Array:
			err = sym.addArrayType(pkg, obj, t, kind, id, n)

		case *types.Slice:
			err = sym.addSliceType(pkg, obj, t, kind, id, n)

		case *types.Map:
			err = sym.addMapType(pkg, obj, t, kind, id, n)

		case *types.Signature:
			err = sym.addSignatureType(pkg, obj, t, kind, id, n)

		case *types.Pointer:
			err = sym.addPointerType(pkg, obj, t, kind, id, n)

		case *types.Interface:
			err = sym.addInterfaceType(pkg, obj, t, kind, id, n)

		case *types.Chan:
			err = fmt.Errorf("gopy: channel type not supported: %s", n)

		default:
			err = fmt.Errorf("unhandled named-type: [%T]\n%#v\n", obj, t)
		}

		if err != nil {
			return err
		}

		// add methods
		for i := 0; i < typ.NumMethods(); i++ {
			m := typ.Method(i)
			if !m.Exported() {
				continue
			}
			if true {
				mid := id + "_" + m.Name()
				mname := m.Name()
				sym.addMethod(pkg, m, m.Type(), skFunc, mid, mname)
			}
		}

	default:
		return fmt.Errorf("unhandled obj [%T]\ntype [%#v]", obj, t)
	}
	return nil
}

func (sym *symtab) addArrayType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	typ := t.Underlying().(*types.Array)
	kind |= skArray
	enam := sym.typename(typ.Elem(), nil)
	elt := sym.sym(enam)
	if elt == nil || elt.goname == "" {
		eltname := sym.typename(typ.Elem(), pkg)
		eobj := sym.pkg.Scope().Lookup(eltname)
		if eobj != nil {
			sym.addSymbol(eobj)
		} else {
			if ntyp, ok := typ.Elem().(*types.Named); ok {
				sym.addType(ntyp.Obj(), typ.Elem())
			}
		}
		elt = sym.sym(enam)
		if elt == nil {
			return fmt.Errorf("gopy: could not retrieve array-elt symbol for %q", enam)
		}
	}
	pyname := n
	pyname = strings.Replace(pyname, "[", "ArrayOf_", 1)
	pyname = strings.Replace(pyname, "]", "_", 1)
	pyname = goToPyName(pyname)
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		pyname:  pyname,
		cgoname: "CGoHandle", // handles
		cpyname: PyHandle,
		pysig:   "[]" + elt.pysig,
		go2py:   "handleFmPtr_" + pyname,
		py2go:   "*ptrFmHandle_" + pyname,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addMapType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	typ := t.Underlying().(*types.Map)
	kind |= skMap
	elt := typ.Elem()
	enam := sym.typename(elt, nil)
	elsym := sym.sym(enam)
	if elsym == nil || elsym.goname == "" {
		eltname, _ := sym.typeNamePkg(elt)
		eobj := sym.pkg.Scope().Lookup(eltname)
		if eobj != nil {
			sym.addSymbol(eobj)
		} else {
			if ntyp, ok := elt.(*types.Named); ok {
				sym.addType(ntyp.Obj(), elt)
			}
		}
		elsym = sym.sym(enam)
		if elsym == nil {
			return fmt.Errorf("gopy: map key type must be named type if outside current package: %q", enam)
		}
	}
	pyname := n
	pyname = strings.Replace(pyname, "[", "MapOf_", 1)
	pyname = strings.Replace(pyname, "]", "_", 1)
	pyname = goToPyName(pyname)
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		pyname:  pyname,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + pyname,
		py2go:   "*ptrFmHandle_" + pyname,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addSliceType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	typ := t.Underlying().(*types.Slice)
	kind |= skSlice
	elt := typ.Elem()
	enam := sym.typename(elt, nil)
	elsym := sym.sym(enam)
	if elsym == nil || elsym.goname == "" {
		eltname, _ := sym.typeNamePkg(elt)
		eobj := sym.pkg.Scope().Lookup(eltname)
		if eobj != nil {
			sym.addSymbol(eobj)
		} else {
			if ntyp, ok := elt.(*types.Named); ok {
				sym.addType(ntyp.Obj(), elt)
			}
		}
		elsym = sym.sym(enam)
		if elsym == nil {
			return fmt.Errorf("gopy: slice type must be named type if outside current package: %q", enam)
		}
		n = "[]" + elsym.goname
	}
	pyname := goToPyName(n)
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		pyname:  pyname,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "[]" + elsym.pysig,
		go2py:   "handleFmPtr_" + pyname,
		py2go:   "*ptrFmHandle_" + pyname,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addStructType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	typ := t.Underlying().(*types.Struct)
	kind |= skStruct
	for i := 0; i < typ.NumFields(); i++ {
		if isPrivate(typ.Field(i).Name()) {
			continue
		}
		ftyp := typ.Field(i).Type()
		fsym := sym.symtype(ftyp)
		if fsym == nil {
			err := sym.addType(typ.Field(i), ftyp)
			if err != nil {
				continue
			}
			fsym = sym.symtype(ftyp)
			if fsym == nil {
				continue
			}
		}
	}
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		pyname:  n,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + n,
		py2go:   "*ptrFmHandle_" + n,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addSignatureType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	//typ := t.(*types.Signature)
	kind |= skSignature
	if (kind & skNamed) == 0 {
		id = hash(id)
	}
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		pyname:  n, // todo: diff
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "callable",
		go2py:   "?",
		py2go:   "?",
	}
	return nil
}

func (sym *symtab) addMethod(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	sig := t.Underlying().(*types.Signature)
	err, _, _ := isPyCompatFunc(sig)
	if err == nil {
		fn := types.ObjectString(obj, nil)
		kind |= skFunc
		sym.syms[fn] = &symbol{
			gopkg:   pkg,
			goobj:   obj,
			gotyp:   t,
			kind:    kind,
			id:      id,
			goname:  n,
			pyname:  n,
			cgoname: fn + "_" + n,
			cpyname: fn + "_" + n,
		}
		sym.processTuple(sig.Results())
		sym.processTuple(sig.Params())
	}
	return nil
}

func (sym *symtab) addPointerType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	typ := t.Underlying().(*types.Pointer)
	etyp := typ.Elem()
	esym := sym.symtype(etyp)
	if esym == nil {
		sym.addType(obj, etyp)
		esym = sym.symtype(etyp)
		if esym == nil {
			return fmt.Errorf("gopy: could not retrieve symbol for %q", sym.typename(etyp, nil))
		}
	}

	pyname := n
	if len(pyname) > 0 && pyname[0] == '*' {
		pyname = pyname[1:]
	}
	pyname = goToPyName(pyname)

	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    esym.kind | skPointer,
		id:      id,
		goname:  n,
		pyname:  pyname,
		cgoname: "CGoHandle", // handles
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + pyname + "_Ptr",
		py2go:   "ptrFmHandle_" + pyname + "_Ptr",
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addInterfaceType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.typename(t, nil)
	typ := t.Underlying().(*types.Interface)
	kind |= skInterface
	// special handling of 'error'
	if isErrorType(typ) {
		return nil
	}

	pyname := goToPyName(n)

	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		pyname:  n,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + pyname,
		py2go:   "ptrFmHandle_" + pyname,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) print() {
	fmt.Printf("\n\n%s\n", strings.Repeat("=", 80))
	for _, n := range sym.names() {
		fmt.Printf("%q (kind=%v)\n", n, sym.syms[n].kind)
	}
	fmt.Printf("%s\n", strings.Repeat("=", 80))
}

func init() {

	look := types.Universe.Lookup
	syms := map[string]*symbol{
		"bool": {
			gopkg:   look("bool").Pkg(),
			goobj:   look("bool"),
			gotyp:   look("bool").Type(),
			kind:    skType | skBasic,
			goname:  "bool",
			pyname:  "bool",
			cgoname: "C.char",
			cpyname: "uint8_t",
			pysig:   "bool",
			go2py:   "boolGoToPy",
			py2go:   "boolPyToGo",
			zval:    "false",
		},

		"byte": {
			gopkg:   look("byte").Pkg(),
			goobj:   look("byte"),
			gotyp:   look("byte").Type(),
			kind:    skType | skBasic,
			goname:  "byte",
			pyname:  "int",
			cpyname: "uint8_t",
			cgoname: "C.char",
			pysig:   "int", // FIXME(sbinet) py2/py3
			go2py:   "C.char",
			py2go:   "byte",
			zval:    "0",
		},

		"int": {
			gopkg:   look("int").Pkg(),
			goobj:   look("int"),
			gotyp:   look("int").Type(),
			kind:    skType | skBasic,
			goname:  "int",
			pyname:  "int",
			cpyname: "int",
			cgoname: "C.long", // see below for 64 bit version
			pysig:   "int",
			go2py:   "C.long",
			py2go:   "int",
			zval:    "0",
		},

		"int8": {
			gopkg:   look("int8").Pkg(),
			goobj:   look("int8"),
			gotyp:   look("int8").Type(),
			kind:    skType | skBasic,
			goname:  "int8",
			pyname:  "int",
			cpyname: "int8_t",
			cgoname: "C.char",
			pysig:   "int",
			go2py:   "C.char",
			py2go:   "int8",
			zval:    "0",
		},

		"int16": {
			gopkg:   look("int16").Pkg(),
			goobj:   look("int16"),
			gotyp:   look("int16").Type(),
			kind:    skType | skBasic,
			goname:  "int16",
			cpyname: "int16_t",
			cgoname: "C.short",
			pysig:   "int",
			go2py:   "C.short",
			py2go:   "int16",
			zval:    "0",
		},

		"int32": {
			gopkg:   look("int32").Pkg(),
			goobj:   look("int32"),
			gotyp:   look("int32").Type(),
			kind:    skType | skBasic,
			goname:  "int32",
			pyname:  "int",
			cpyname: "int32_t",
			cgoname: "C.long",
			pysig:   "int",
			go2py:   "C.long",
			py2go:   "int32",
			zval:    "0",
		},

		"int64": {
			gopkg:   look("int64").Pkg(),
			goobj:   look("int64"),
			gotyp:   look("int64").Type(),
			kind:    skType | skBasic,
			goname:  "int64",
			pyname:  "int",
			cpyname: "int64_t",
			cgoname: "C.longlong",
			pysig:   "long",
			go2py:   "C.longlong",
			py2go:   "int64",
			zval:    "0",
		},

		"uint": {
			gopkg:   look("uint").Pkg(),
			goobj:   look("uint"),
			gotyp:   look("uint").Type(),
			kind:    skType | skBasic,
			goname:  "uint",
			pyname:  "int",
			cpyname: "unsigned int",
			cgoname: "C.uint",
			pysig:   "int",
			go2py:   "C.uint",
			py2go:   "uint",
			zval:    "0",
		},

		"uint8": {
			gopkg:   look("uint8").Pkg(),
			goobj:   look("uint8"),
			gotyp:   look("uint8").Type(),
			kind:    skType | skBasic,
			goname:  "uint8",
			pyname:  "int",
			cpyname: "uint8_t",
			cgoname: "C.uchar",
			pysig:   "int",
			go2py:   "C.uchar",
			py2go:   "uint8",
			zval:    "0",
		},

		"uint16": {
			gopkg:   look("uint16").Pkg(),
			goobj:   look("uint16"),
			gotyp:   look("uint16").Type(),
			kind:    skType | skBasic,
			goname:  "uint16",
			pyname:  "int",
			cpyname: "uint16_t",
			cgoname: "C.ushort",
			pysig:   "int",
			go2py:   "C.ushort",
			py2go:   "uint16",
			zval:    "0",
		},

		"uint32": {
			gopkg:   look("uint32").Pkg(),
			goobj:   look("uint32"),
			gotyp:   look("uint32").Type(),
			kind:    skType | skBasic,
			goname:  "uint32",
			pyname:  "long",
			cpyname: "uint32_t",
			cgoname: "C.ulong",
			pysig:   "long",
			go2py:   "C.ulong",
			py2go:   "uint32",
			zval:    "0",
		},

		"uint64": {
			gopkg:   look("uint64").Pkg(),
			goobj:   look("uint64"),
			gotyp:   look("uint64").Type(),
			kind:    skType | skBasic,
			goname:  "uint64",
			pyname:  "long",
			cpyname: "uint64_t",
			cgoname: "C.ulonglong",
			pysig:   "long",
			go2py:   "C.ulonglong",
			py2go:   "uint64",
			zval:    "0",
		},

		"float32": {
			gopkg:   look("float32").Pkg(),
			goobj:   look("float32"),
			gotyp:   look("float32").Type(),
			kind:    skType | skBasic,
			goname:  "float32",
			pyname:  "float",
			cpyname: "float",
			cgoname: "C.float",
			pysig:   "float",
			go2py:   "C.float",
			py2go:   "float32",
			zval:    "0",
		},

		"float64": {
			gopkg:   look("float64").Pkg(),
			goobj:   look("float64"),
			gotyp:   look("float64").Type(),
			kind:    skType | skBasic,
			goname:  "float64",
			pyname:  "float",
			cpyname: "double",
			cgoname: "C.double",
			pysig:   "float",
			go2py:   "C.double",
			py2go:   "float64",
			zval:    "0",
		},

		"complex64": {
			gopkg:   look("complex64").Pkg(),
			goobj:   look("complex64"),
			gotyp:   look("complex64").Type(),
			kind:    skType | skBasic,
			goname:  "complex64",
			pyname:  "complex",
			cpyname: "float complex",
			cgoname: "GoComplex64",
			pysig:   "complex",
			go2py:   "cgopy_cnv_go2py_complex64",
			py2go:   "cgopy_cnv_py2go_complex64",
			zval:    "0",
		},

		"complex128": {
			gopkg:   look("complex128").Pkg(),
			goobj:   look("complex128"),
			gotyp:   look("complex128").Type(),
			kind:    skType | skBasic,
			goname:  "complex128",
			pyname:  "complex",
			cpyname: "double complex",
			cgoname: "GoComplex128",
			pysig:   "complex",
			go2py:   "cgopy_cnv_go2py_complex128",
			py2go:   "cgopy_cnv_py2go_complex128",
			zval:    "0",
		},

		"string": {
			gopkg:   look("string").Pkg(),
			goobj:   look("string"),
			gotyp:   look("string").Type(),
			kind:    skType | skBasic,
			goname:  "string",
			pyname:  "str",
			cpyname: "char*",
			cgoname: "*C.char",
			pysig:   "str",
			go2py:   "C.CString",
			py2go:   "C.GoString",
			zval:    `""`,
		},

		"rune": { // FIXME(sbinet) py2/py3
			gopkg:   look("rune").Pkg(),
			goobj:   look("rune"),
			gotyp:   look("rune").Type(),
			kind:    skType | skBasic,
			goname:  "rune",
			pyname:  "str",
			cpyname: "int32_t",
			cgoname: "C.long",
			pysig:   "str",
			go2py:   "C.long",
			py2go:   "rune",
			zval:    "0",
		},

		"error": {
			gopkg:   look("error").Pkg(),
			goobj:   look("error"),
			gotyp:   look("error").Type(),
			kind:    skType | skInterface,
			goname:  "error",
			pyname:  "str",
			cpyname: "char*",
			cgoname: "*C.char",
			pysig:   "str",
			go2py:   "C.CString",
			py2go:   "C.GoString",
			zval:    `""`,
		},
	}

	if reflect.TypeOf(int(0)).Size() == 8 {
		syms["int"] = &symbol{
			gopkg:   look("int").Pkg(),
			goobj:   look("int"),
			gotyp:   look("int").Type(),
			kind:    skType | skBasic,
			goname:  "int",
			pyname:  "int",
			cpyname: "int64_t",
			cgoname: "C.longlong",
			pysig:   "int",
			go2py:   "C.longlong",
			py2go:   "int",
			zval:    "0",
		}
		syms["uint"] = &symbol{
			gopkg:   look("uint").Pkg(),
			goobj:   look("uint"),
			gotyp:   look("uint").Type(),
			kind:    skType | skBasic,
			goname:  "uint",
			pyname:  "int",
			cpyname: "uint64_t",
			cgoname: "C.ulonglong",
			pysig:   "int",
			go2py:   "C.ulonglong",
			py2go:   "uint",
			zval:    "0",
		}
	}

	for _, o := range []struct {
		kind  types.BasicKind
		tname string
		uname string
	}{
		{types.UntypedBool, "bool", "bool"},
		{types.UntypedInt, "int", "int"},
		{types.UntypedRune, "rune", "rune"},
		{types.UntypedFloat, "float64", "float"},
		{types.UntypedComplex, "complex128", "complex"},
		{types.UntypedString, "string", "string"},
		//FIXME(sbinet): what should be the python equivalent?
		//{types.UntypedNil, "nil", "nil"},
	} {
		sym := *syms[o.tname]
		n := "untyped " + o.uname
		syms[n] = &sym
	}

	universe = &symtab{
		pkg:    nil,
		syms:   syms,
		parent: nil,
	}
}
