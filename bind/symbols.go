// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"hash/fnv"
	"sort"
	"strings"
	"sync"
)

var (
	universeMutex sync.Mutex
	universe      *symtab // universe contains Go global types that are not genrated
	current       *symtab // current contains all symbols from current multi-package gen run
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

var pyKeywords = map[string]struct{}{
	"False": struct{}{}, "None": struct{}{}, "True": struct{}{}, "and": struct{}{}, "as": struct{}{}, "assert": struct{}{}, "break": struct{}{}, "class": struct{}{}, "continue": struct{}{}, "def": struct{}{}, "del": struct{}{}, "elif": struct{}{}, "else": struct{}{}, "except": struct{}{}, "finally": struct{}{}, "for": struct{}{}, "from": struct{}{}, "global": struct{}{}, "if": struct{}{}, "import": struct{}{}, "in": struct{}{}, "is": struct{}{}, "lambda": struct{}{}, "nonlocal": struct{}{}, "not": struct{}{}, "or": struct{}{}, "pass": struct{}{}, "raise": struct{}{}, "return": struct{}{}, "try": struct{}{}, "while": struct{}{}, "with": struct{}{}, "yield": struct{}{},
}

// pySafeName returns a name that python will not barf on
func pySafeName(nm string) string {
	if _, bad := pyKeywords[nm]; bad {
		return "my" + nm
	}
	return nm
}

// pySafeArg returns an arg name that python will not barf on
func pySafeArg(anm string, idx int) string {
	if anm == "" {
		anm = fmt.Sprintf("arg_%d", idx)
	}
	return pySafeName(anm)
}

// isPyCompatVar checks if var is compatible with python
func isPyCompatVar(v *symbol) error {
	if v == nil {
		return fmt.Errorf("gopy: var symbol not found")
	}
	if v.isPointer() && v.isBasic() {
		return fmt.Errorf("gopy: var is pointer to basic type")
	}
	if isErrorType(v.gotyp) {
		return fmt.Errorf("gopy: var is error type")
	}
	if v.gotyp.String() == "interface{}" {
		return fmt.Errorf("gopy: var is interface{}")
	}
	if _, isChan := v.gotyp.(*types.Chan); isChan {
		return fmt.Errorf("gopy: var is channel type")
	}
	return nil
}

// isPyCompatType checks if type is compatible with python
func isPyCompatType(typ types.Type) error {
	typ = typ.Underlying()
	if ptyp, isPtr := typ.(*types.Pointer); isPtr {
		if _, isBasic := ptyp.Elem().(*types.Basic); isBasic {
			return fmt.Errorf("gopy: type is pointer to basic type")
		}
	}
	if isErrorType(typ) {
		return fmt.Errorf("gopy: type is error type")
	}
	if _, isChan := typ.(*types.Chan); isChan {
		return fmt.Errorf("gopy: type is channel type")
	}
	if typ.String() == "interface{}" {
		return fmt.Errorf("gopy: type is interface{}")
	}
	return nil
}

// isPyCompatField checks if field is compatible with python
func isPyCompatField(f *types.Var) (error, *symbol) {
	if !f.Exported() || f.Embedded() {
		return fmt.Errorf("gopy: field not exported or is embedded"), nil
	}
	ftyp := current.symtype(f.Type())
	if _, isSig := f.Type().Underlying().(*types.Signature); isSig {
		return fmt.Errorf("gopy: type is function signature"), nil
	}
	return isPyCompatVar(ftyp), ftyp
}

// isPyCompatFunc checks if function signature is a python-compatible function.
// Returns nil if function is compatible, err message if not.
// Also returns the return type of the function
// haserr is true if 2nd arg is an error type, which is only
// supported form of multi-return-value functions
// hasfun is true if one of the args is a function signature
func isPyCompatFunc(sig *types.Signature) (err error, ret types.Type, haserr, hasfun bool) {
	res := sig.Results()

	if sig.Variadic() {
		err = fmt.Errorf("gopy: not yet supporting variadic functions: %s", sig.String())
		return
	}

	switch res.Len() {
	case 2:
		if !isErrorType(res.At(1).Type()) {
			err = fmt.Errorf("gopy: second result value must be of type error: %s", sig.String())
			return
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
		err = fmt.Errorf("gopy: too many results to return: %s", sig.String())
		return
	}

	if ret != nil {
		if err = isPyCompatType(ret); err != nil {
			return
		}
		if _, isSig := ret.Underlying().(*types.Signature); isSig {
			return
		}
	}

	args := sig.Params()
	nargs := args.Len()
	for i := 0; i < nargs; i++ {
		arg := args.At(i)
		argt := arg.Type()
		if err = isPyCompatType(argt); err != nil {
			return
		}
		if _, isSig := argt.Underlying().(*types.Signature); isSig {
			if !hasfun {
				hasfun = true
			} else {
				err = fmt.Errorf("gopy: only one function signature arg allowed: %s", sig.String())
				return
			}
		}
	}
	return
}

// symbol is an exported symbol in a go package
type symbol struct {
	kind         symkind
	gopkg        *types.Package
	goobj        types.Object
	gotyp        types.Type
	doc          string
	id           string // canonical name for referring to the symbol as a valid variable / function name etc
	goname       string // go formatted type name (gofmt)
	cgoname      string // type name of entity for cgo -- handle for objs
	cpyname      string // type name of entity for cgo - python -- handle..
	pysig        string // type string for doc-signatures
	go2py        string // name of go->py converter function
	go2pyParenEx string // extra parentheses needed at end of go2py (beyond 1 default)
	py2go        string // name of py->go converter function
	py2goParenEx string // extra parentheses needed at end of py2go (beyond 1 default)
	zval         string // zero value representation
	pyfmt        string // python Py_BuildValue / Py_ParseTuple formatting string
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
	return !s.isBasic() && !s.isSignature()
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

// gofmt returns the type name that is always qualified by an appropriate package name
// this should always be used for "goname" in general.
func (s *symbol) gofmt() string {
	return types.TypeString(s.GoType(), func(pkg *types.Package) string { return pkg.Name() })
}

// idname returns gofmt with . -> _ -- this should always be used for id
func (s *symbol) idname() string {
	return strings.Replace(s.gofmt(), ".", "_", -1)
}

// pyPkgId returns the python package-qualified version of Id
func (s *symbol) pyPkgId(curPkg *types.Package) string {
	pnm := s.gopkg.Name()
	if _, has := thePyGen.pkgmap[s.gopkg.Path()]; !has { // external symbols are all in go package
		if pnm == "go" {
			return s.id
		} else {
			return "go." + s.id
		}
	}
	if pnm == "go" {
		return pnm + "." + s.id
	}
	if !s.isNamed() && (s.isMap() || s.isSlice() || s.isArray()) {
		//		idnm := strings.TrimPrefix(s.id[uidx+1:], pnm+"_") // in case it has that redundantly
		if s.gopkg.Path() != curPkg.Path() {
			return pnm + "." + s.id
		} else {
			return s.id
		}
	}
	uidx := strings.Index(s.id, "_")
	if uidx < 0 {
		return s.id // shouldn't happen
	}
	idnm := strings.TrimPrefix(s.id[uidx+1:], pnm+"_") // in case it has that redundantly
	if s.gopkg.Path() != curPkg.Path() {
		return pnm + "." + idnm
	} else {
		return idnm
	}
}

//////////////////////////////////////////////////////////////
// symtab

// symtab is a table of symbols in a go package
type symtab struct {
	pkg     *types.Package // current package only -- can change depending..
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

func (sym *symtab) addImport(pkg *types.Package) {
	p := pkg.Path()
	sym.imports[p] = p
}

func (sym *symtab) typeof(n string) *symbol {
	s := sym.sym(n)
	switch s.kind {
	case skVar, skConst:
		tname := sym.fullTypeString(s.goobj.Type())
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

// fullTypeString returns the fully-qualified type string with entire package import path
func (sym *symtab) fullTypeString(t types.Type) string {
	return types.TypeString(t, nil)
}

func (sym *symtab) symtype(t types.Type) *symbol {
	tname := sym.fullTypeString(t)
	return sym.sym(tname)
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

// typeGoName returns the go type name that is always qualified by an appropriate package name
// this should always be used for "goname" in general.
func typeGoName(t types.Type) string {
	return types.TypeString(t, func(pkg *types.Package) string { return pkg.Name() })
}

// typeIdName returns typeGoName with . -> _ -- this should always be used for id
func typeIdName(t types.Type) string {
	idn := strings.Replace(typeGoName(t), ".", "_", -1)
	if _, isary := t.(*types.Array); isary {
		idn = strings.Replace(idn, "[", "Array_", 1)
		idn = strings.Replace(idn, "]", "_", 1)
	}
	idn = strings.Replace(idn, "[]", "Slice_", -1)
	idn = strings.Replace(idn, "[", "Map_", -1)
	idn = strings.Replace(idn, "]", "_", -1)
	idn = strings.Replace(idn, "{}", "_", -1)
	idn = strings.Replace(idn, "*", "Ptr_", -1)
	return idn
}

func (sym *symtab) addSymbol(obj types.Object) error {
	fn := types.ObjectString(obj, nil)
	n := obj.Name()
	pkg := obj.Pkg()
	id := n
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
			cgoname: n, // todo: type names!
			cpyname: n,
		}
		tsym := sym.symtype(obj.Type())
		if tsym == nil {
			return sym.addType(obj, obj.Type())
		}

	case *types.Var:
		sym.syms[fn] = &symbol{
			gopkg:   pkg,
			goobj:   obj,
			kind:    skVar,
			id:      id,
			goname:  n,
			cgoname: n,
			cpyname: n,
		}
		tsym := sym.symtype(obj.Type())
		if tsym == nil {
			return sym.addType(obj, obj.Type())
		}

	case *types.Func:
		sig := obj.Type().Underlying().(*types.Signature)
		err, _, _, _ := isPyCompatFunc(sig)
		if err == nil {
			sym.syms[fn] = &symbol{
				gopkg:   pkg,
				goobj:   obj,
				kind:    skFunc,
				id:      id,
				goname:  n,
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

// processTuple ensures that all types in a tuple are in sym table
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

// buildTuple returns a string of Go code that builds a PyTuple
// for the given tuple (e.g., function args).
// varnm is the name of the local variable for the built tuple
func (sym *symtab) buildTuple(tuple *types.Tuple, varnm string) (string, error) {
	sz := tuple.Len()
	if sz == 0 {
		return "", fmt.Errorf("buildTuple: no elements")
	}
	bstr := fmt.Sprintf("%s := C.PyTuple_New(%d); ", varnm, sz)
	for i := 0; i < sz; i++ {
		v := tuple.At(i)
		typ := v.Type()
		anm := pySafeArg(v.Name(), i)
		vsym := sym.symtype(typ)
		if vsym == nil {
			err := sym.addType(v, typ)
			if err != nil {
				return "", err
			}
			vsym = sym.symtype(typ)
			if vsym == nil {
				return "", fmt.Errorf("buildTuple: type still not found: %s", typ.String())
			}
		}
		bt, isb := typ.Underlying().(*types.Basic)
		switch {
		case vsym.hasHandle(): // note: assuming int64 handles
			bstr += fmt.Sprintf("C.PyTuple_SetItem(%s, %d, C.gopy_build_int64(C.int64_t(%s(%s)%s))); ", varnm, i, vsym.go2py, anm, vsym.go2pyParenEx)
		case isb:
			bk := bt.Kind()
			switch {
			case bk >= types.Int && bk <= types.Int64:
				bstr += fmt.Sprintf("C.PyTuple_SetItem(%s, %d, C.gopy_build_int64(C.int64_t(%s))); ", varnm, i, anm)
			case bk >= types.Uint && bk <= types.Uintptr:
				bstr += fmt.Sprintf("C.PyTuple_SetItem(%s, %d, C.gopy_build_uint64(C.uint64_t(%s))); ", varnm, i, anm)
			case bk >= types.Float32 && bk <= types.Float64:
				bstr += fmt.Sprintf("C.PyTuple_SetItem(%s, %d, C.gopy_build_float64(C.double(%s))); ", varnm, i, anm)
			case bk == types.String:
				bstr += fmt.Sprintf("C.PyTuple_SetItem(%s, %d, C.gopy_build_string(C.CString(%s))); ", varnm, i, anm)
			case bk == types.Bool:
				bstr += fmt.Sprintf("C.PyTuple_SetItem(%s, %d, C.gopy_build_bool(C.uint8_t(boolGoToPy(%s)))); ", varnm, i, anm)
			}
		default:
			return "", fmt.Errorf("buildTuple: type not handled: %s", typ.String())
		}
	}
	return bstr, nil
}

// typeNamePkg gets the goname and package for a given types.Type -- deals with
// naming for types in other packages, and adds those packages to imports paths.
// Falls back on sym.pkg if no other package info avail.
func (sym *symtab) typeNamePkg(t types.Type) (gonm, idnm string, pkg *types.Package) {
	gonm = typeGoName(t)
	idnm = typeIdName(t)
	pkg = typePkg(t)
	if pkg != nil {
		sym.addImport(pkg)
	}
	if pkg == nil {
		pkg = sym.pkg
	}
	return
}

func (sym *symtab) addType(obj types.Object, t types.Type) error {
	fn := sym.fullTypeString(t)
	n, id, pkg := sym.typeNamePkg(t)
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
			py2go := typeGoName(t)
			py2goParEx := ""
			if styp.py2go != "" {
				py2go += "(" + styp.py2go
				py2goParEx = ")"
			}
			go2py := styp.goname
			go2pyParEx := ""
			if styp.go2py != "" {
				go2py = styp.go2py + "(" + go2py
				go2pyParEx = ")"
			}

			sym.syms[fn] = &symbol{
				gopkg:        pkg,
				goobj:        obj,
				gotyp:        t,
				kind:         kind | skBasic,
				id:           id,
				goname:       styp.goname,
				cgoname:      styp.cgoname,
				cpyname:      styp.cpyname,
				pysig:        styp.pysig,
				go2py:        go2py,
				go2pyParenEx: go2pyParEx,
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
			mid := id + "_" + m.Name()
			mname := m.Name()
			sym.addMethod(pkg, m, m.Type(), skFunc, mid, mname)
		}

	default:
		return fmt.Errorf("unhandled obj [%T]\ntype [%#v]", obj, t)
	}
	return nil
}

func (sym *symtab) addArrayType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	typ := t.Underlying().(*types.Array)
	kind |= skArray
	elt := typ.Elem()
	enam := sym.fullTypeString(elt)
	elsym := sym.sym(enam)
	if elsym == nil || elsym.goname == "" {
		eltname, _, _ := sym.typeNamePkg(elt)
		eobj := sym.pkg.Scope().Lookup(eltname)
		if eobj != nil {
			sym.addSymbol(eobj)
		} else {
			if ntyp, ok := typ.Elem().(*types.Named); ok {
				sym.addType(ntyp.Obj(), elt)
			}
		}
		elsym = sym.sym(enam)
		if elsym == nil {
			return fmt.Errorf("gopy: could not retrieve array-elt symbol for %q", enam)
		}
	}
	if elsym.isSignature() {
		return fmt.Errorf("gopy: array value type cannot be signature / func: %q", enam)
	}
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		cgoname: "CGoHandle", // handles
		cpyname: PyHandle,
		pysig:   "[]" + elsym.pysig,
		go2py:   "handleFmPtr_" + id,
		py2go:   "*ptrFmHandle_" + id,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addMapType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	typ := t.Underlying().(*types.Map)
	kind |= skMap
	elt := typ.Elem()
	enam := sym.fullTypeString(elt)
	elsym := sym.sym(enam)
	if elsym == nil || elsym.goname == "" {
		eltname, _, _ := sym.typeNamePkg(elt)
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
			return fmt.Errorf("gopy: map value type must be named type if outside current package: %q", enam)
		}
	}
	if elsym.isSignature() {
		return fmt.Errorf("gopy: map value type cannot be signature / func: %q", enam)
	}
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + id,
		py2go:   "*ptrFmHandle_" + id,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addSliceType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	typ := t.Underlying().(*types.Slice)
	kind |= skSlice
	elt := typ.Elem()
	enam := sym.fullTypeString(elt)
	elsym := sym.sym(enam)
	if elsym == nil || elsym.goname == "" {
		eltname, _, _ := sym.typeNamePkg(elt)
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
	if elsym.isSignature() {
		return fmt.Errorf("gopy: slice value type cannot be signature / func: %q", enam)
	}
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "[]" + elsym.pysig,
		go2py:   "handleFmPtr_" + id,
		py2go:   "*ptrFmHandle_" + id,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addStructType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	typ := t.Underlying().(*types.Struct)
	kind |= skStruct
	// add our type first before adding fields -- prevents loops!
	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + id,
		py2go:   "*ptrFmHandle_" + id,
		zval:    "nil",
	}
	for i := 0; i < typ.NumFields(); i++ {
		if isPrivate(typ.Field(i).Name()) {
			continue
		}
		f := typ.Field(i)
		if !f.Exported() || f.Embedded() {
			continue
		}
		ftyp := f.Type()
		fsym := sym.symtype(ftyp)
		if fsym == nil {
			err := sym.addType(f, ftyp)
			if err != nil {
				continue
			}
			fsym = sym.symtype(ftyp)
			if fsym == nil {
				continue
			}
		}
	}
	return nil
}

func (sym *symtab) addSignatureType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	kind |= skSignature

	nsig := typeGoName(t.Underlying())
	py2g := fmt.Sprintf("%s { ", nsig)

	sig := t.Underlying().(*types.Signature)
	args := sig.Params()
	rets := sig.Results()
	if rets.Len() > 0 { // todo 1
		return fmt.Errorf("multiple return values not supported")
	}
	retstr := ""
	if rets.Len() == 1 {
		retstr = "_fcret := "
	}

	py2g += "if C.PyCallable_Check(_fun_arg) == 0 { return }\n"
	py2g += "_gstate := C.PyGILState_Ensure()\n"
	nargs := args.Len()
	if nargs > 0 {
		bstr, err := sym.buildTuple(args, "_fcargs")
		if err != nil {
			return err
		}
		py2g += bstr + retstr
		py2g += fmt.Sprintf("C.PyObject_CallObject(_fun_arg, _fcargs)\n")
		py2g += "C.gopy_decref(_fcargs)\n"
	} else {
		py2g += retstr + "C.PyObject_CallObject(_fun_arg, nil)\n"
	}
	py2g += "C.gopy_err_handle()\n"
	py2g += "C.PyGILState_Release(_gstate)\n"
	if rets.Len() == 1 {
		ret := rets.At(0)
		rsym := sym.symtype(ret.Type())
		if rsym == nil {
			return fmt.Errorf("return type not supported: %s", n)
		}
		if rsym.py2go != "" {
			py2g += fmt.Sprintf("return %s(_fcret)%s", rsym.py2go, rsym.py2goParenEx)
		} else {
			py2g += "return _fcret"
		}
	}
	py2g += "}"

	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		cgoname: "*C.PyObject",
		cpyname: "PyObject*",
		pysig:   "callable",
		go2py:   "?",
		py2go:   py2g,
	}
	return nil
}

func (sym *symtab) addMethod(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	sig := t.Underlying().(*types.Signature)
	err, _, _, _ := isPyCompatFunc(sig)
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
			cgoname: fn + "_" + id,
			cpyname: fn + "_" + id,
		}
		sym.processTuple(sig.Results())
		sym.processTuple(sig.Params())
	}
	return nil
}

func (sym *symtab) addPointerType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	typ := t.Underlying().(*types.Pointer)
	etyp := typ.Elem()
	esym := sym.symtype(etyp)
	if esym == nil {
		sym.addType(obj, etyp)
		esym = sym.symtype(etyp)
		if esym == nil {
			return fmt.Errorf("gopy: could not retrieve symbol for %q", sym.fullTypeString(etyp))
		}
	}

	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    esym.kind | skPointer,
		id:      id,
		goname:  n,
		cgoname: "CGoHandle", // handles
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + id,
		py2go:   "ptrFmHandle_" + id,
		zval:    "nil",
	}
	return nil
}

func (sym *symtab) addInterfaceType(pkg *types.Package, obj types.Object, t types.Type, kind symkind, id, n string) error {
	fn := sym.fullTypeString(t)
	typ := t.Underlying().(*types.Interface)
	kind |= skInterface
	// special handling of 'error'
	if isErrorType(typ) {
		return nil
	}

	sym.syms[fn] = &symbol{
		gopkg:   pkg,
		goobj:   obj,
		gotyp:   t,
		kind:    kind,
		id:      id,
		goname:  n,
		cgoname: "CGoHandle",
		cpyname: PyHandle,
		pysig:   "object",
		go2py:   "handleFmPtr_" + id,
		py2go:   "ptrFmHandle_" + id,
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

	universe = newSymtab(nil, nil)
	universe.parent = nil

	universe.syms = stdBasicTypes()

	addStdSliceMaps()

	current = newSymtab(nil, universe)
}
