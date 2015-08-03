// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"golang.org/x/tools/go/types"
)

// symbol is an exported symbol in a go package
type symbol struct {
	obj     types.Object
	goname  string
	cgoname string
	cpyname string
}

// symtab is a table of symbols in a go package
type symtab struct {
	syms map[string]symbol
}

func newSymtab() symtab {
	return symtab{
		syms: make(map[string]symbol),
	}
}

func (sym *symtab) addSymbol(obj types.Object) {
	n := obj.Name()
	pkg := obj.Pkg()
	id := pkg.Name() + "_" + n
	switch obj.(type) {
	case *types.Const:
		sym.syms[n] = symbol{
			obj:     obj,
			goname:  n,
			cgoname: "cgopy_const_" + id,
			cpyname: "cgopy_const_" + id,
		}

	case *types.Var:
		sym.syms[n] = symbol{
			obj:     obj,
			goname:  n,
			cgoname: "cgopy_var_" + id,
			cpyname: "cgopy_var_" + id,
		}
		sym.addType(obj, obj.Type())

	case *types.Func:
		sym.syms[n] = symbol{
			obj:     obj,
			goname:  n,
			cgoname: "cgopy_func_" + id,
			cpyname: "cgopy_func_" + id,
		}

	case *types.TypeName:
		sym.addType(obj, obj.Type())
	}
}

func (sym *symtab) addType(obj types.Object, t types.Type) {
	n := obj.Name()
	pkg := obj.Pkg()
	id := pkg.Name() + "_" + n
	switch typ := obj.(type) {
	case *types.TypeName:
		named := typ.Type().(*types.Named)
		switch named.Underlying().(type) {
		case *types.Struct:
			sym.syms[n] = symbol{
				obj:     obj,
				goname:  n,
				cgoname: "cgopy_type_" + id,
				cpyname: "cgopy_type_" + id,
			}
		}
	}
}
