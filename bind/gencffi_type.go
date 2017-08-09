// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
)

// genType generates a Python class wrapping the given Go type.
func (g *cffiGen) genType(sym *symbol) {
	if !sym.isType() {
		return
	}
	if sym.isStruct() {
		return
	}
	if sym.isBasic() && !sym.isNamed() {
		return
	}

	var typename string

	if sym.isNamed() {
		typename = sym.goname
	} else {
		typename = sym.cgoname
	}

	pkgname := g.pkg.pkg.Name()
	g.wrapper.Printf(`
# Python type for %[1]s.%[2]s
class %[2]s(object):
    ""%[3]q""
`, pkgname,
		typename,
		sym.doc,
	)
	g.wrapper.Indent()
	g.genTypeInit(sym)
	g.genTypeMethod(sym)
	g.genTypeTPRepr(sym)
	g.genTypeTPStr(sym)

	if sym.isPySequence() {
		g.genTypeLen(sym)
		g.genTypeGetItem(sym)
		g.genTypeSetItem(sym)
	}

	if sym.isSlice() {
		g.genTypeIAdd(sym)
	}

	g.wrapper.Outdent()
}

// genTypeIAdd generates Type __iadd__.
func (g *cffiGen) genTypeIAdd(sym *symbol) {
	g.wrapper.Printf("def __iadd__(self, value):\n")
	g.wrapper.Indent()
	switch {
	case sym.isSlice():
		g.wrapper.Printf("if not isinstance(value, collections.Iterable):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__iadd__ takes a iterable type as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Slice)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("for elt in value:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(elt)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_append(self.cgopy, pyitem)\n", sym.id)
		g.wrapper.Outdent()
	default:
		panic(fmt.Errorf(
			"gopy: __iadd__ for %s not handled",
			sym.gofmt(),
		))
	}
	g.wrapper.Printf("return self\n")
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeGetItem generates __getitem__ of a type.
func (g *cffiGen) genTypeGetItem(sym *symbol) {
	g.wrapper.Printf("def __getitem__(self, idx):\n")
	g.wrapper.Indent()
	switch {
	case sym.isArray():
		g.wrapper.Printf("if idx >= len(self):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise IndexError('array index out of range')\n")
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Array)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("item = _cffi_helper.lib.cgo_func_%[1]s_item(self.cgopy, idx)\n", sym.id)
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(item)\n", esym.c2py)
	case sym.isSlice():
		g.wrapper.Printf("if idx >= len(self):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise IndexError('slice index out of range')\n")
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Slice)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("item = _cffi_helper.lib.cgo_func_%[1]s_item(self.cgopy, idx)\n", sym.id)
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(item)\n", esym.c2py)
	case sym.isMap():
		typ := sym.GoType().Underlying().(*types.Map)
		ksym := g.pkg.syms.symtype(typ.Key())
		vsym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("pykey = _cffi_helper.cffi_%[1]s(idx)\n", ksym.py2c)
		g.wrapper.Printf("item = _cffi_helper.lib.cgo_func_%[1]s_get(self.cgopy, pykey)\n", sym.id)
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(item)\n", vsym.c2py)
	default:
		panic(fmt.Errorf(
			"gopy: __getitem__ for %s not handled",
			sym.gofmt(),
		))
	}
	g.wrapper.Printf("return pyitem\n")
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeInit generates __init__ of a type.
func (g *cffiGen) genTypeInit(sym *symbol) {
	nargs := 1
	g.wrapper.Printf("def __init__(self, *args, **kwargs):\n")
	g.wrapper.Indent()
	g.wrapper.Printf("nkwds = len(kwargs)\n")
	g.wrapper.Printf("nargs = len(args)\n")
	g.wrapper.Printf("if nkwds + nargs > %[1]d:\n", nargs)
	g.wrapper.Indent()
	g.wrapper.Printf("raise TypeError('%s.__init__ takes at most %d argument(s)')\n",
		sym.goname,
		nargs,
	)
	g.wrapper.Outdent()
	g.wrapper.Printf("self.cgopy = _cffi_helper.lib.cgo_func_%[1]s_new()\n", sym.id)
	switch {
	case sym.isBasic():
		bsym := g.pkg.syms.symtype(sym.GoType().Underlying())
		g.wrapper.Printf("if len(args) == 1:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("self.cgopy = _cffi_helper.cffi_%[1]s(args[0])\n",
			bsym.py2c,
		)
		g.wrapper.Outdent()
	case sym.isArray():
		g.wrapper.Printf("if args:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("if not isinstance(args[0], collections.Sequence):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__init__ takes a sequence as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Array)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("if len(args[0]) > %[1]d:\n", typ.Len())
		g.wrapper.Indent()
		g.wrapper.Printf("raise ValueError('%[1]s.__init__ takes a sequence of size at most %[2]d')\n",
			sym.goname,
			typ.Len(),
		)
		g.wrapper.Outdent()
		g.wrapper.Printf("for idx, elt in enumerate(args[0]):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(elt)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_ass_item(self.cgopy, idx, pyitem)\n", sym.id)
		g.wrapper.Outdent()
		g.wrapper.Outdent()
	case sym.isSlice():
		g.wrapper.Printf("if args:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("if not isinstance(args[0], collections.Iterable):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__init__ takes a sequence as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Slice)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("for elt in args[0]:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(elt)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_append(self.cgopy, pyitem)\n", sym.id)
		g.wrapper.Outdent()
		g.wrapper.Outdent()
	case sym.isMap():
		g.wrapper.Printf("if args:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("if not isinstance(args[0], collections.Mapping):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__init__ takes a mapping as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Map)
		esym := g.pkg.syms.symtype(typ.Elem())
		ksym := g.pkg.syms.symtype(typ.Key())
		g.wrapper.Printf("for k, v in args[0].items():\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pykey = _cffi_helper.cffi_%[1]s(k)\n", ksym.py2c)
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(v)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_set(self.cgopy, pykey, pyitem)\n", sym.id)
		g.wrapper.Outdent()
		g.wrapper.Outdent()
	case sym.isSignature():
		//TODO(corona10)
	case sym.isInterface():
		//TODO(corona10)
	default:
		panic(fmt.Errorf(
			"gopy: init for %s not handled",
			sym.gofmt(),
		))
	}
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeConverter generates a Type converter between C and Python.
func (g *cffiGen) genTypeLen(sym *symbol) {
	g.wrapper.Printf("def __len__(self):\n")
	g.wrapper.Indent()
	switch {
	case sym.isArray():
		typ := sym.GoType().Underlying().(*types.Array)
		g.wrapper.Printf("return %[1]d\n", typ.Len())
	case sym.isSlice():
		g.wrapper.Printf("return ffi.cast('GoSlice*', self.cgopy).len\n")
	case sym.isMap():
		g.wrapper.Printf("return _cffi_helper.lib.cgo_func_%[1]s_len(self.cgopy)\n", sym.id)
	default:
		panic(fmt.Errorf(
			"gopy: len for %s not handled",
			sym.gofmt(),
		))
	}
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeConverter generates Type converter between C and Python.
func (g *cffiGen) genTypeConverter(sym *symbol) {
	var typename string
	if sym.isNamed() {
		typename = sym.goname
	} else {
		typename = sym.cgoname
	}
	g.wrapper.Printf("# converters for %s - %s\n", sym.id, sym.goname)
	g.wrapper.Printf("@staticmethod\n")
	g.wrapper.Printf("def cffi_cgopy_cnv_py2c_%[1]s(o):\n", sym.id)
	g.wrapper.Indent()
	g.wrapper.Printf("if type(o) is %[1]s:\n", typename)
	g.wrapper.Indent()
	g.wrapper.Printf("return o.cgopy\n")
	g.wrapper.Outdent()
	switch {
	case sym.isBasic():
		g.wrapper.Printf("return _cffi_helper.cffi_cgopy_cnv_py2c_%[1]s(o)\n", sym.goname)
	case sym.isArray():
		g.wrapper.Printf("if not isinstance(o, collections.Iterable):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__init__ takes a sequence as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Array)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("if len(o) > %[1]d:\n", typ.Len())
		g.wrapper.Indent()
		g.wrapper.Printf("raise ValueError('%[1]s.__init__ takes a sequence of size at most %[2]d')\n",
			sym.goname,
			typ.Len(),
		)
		g.wrapper.Outdent()
		g.wrapper.Printf("c = _cffi_helper.lib.cgo_func_%[1]s_new()\n", sym.id)
		g.wrapper.Printf("for idx, elt in enumerate(o):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(elt)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_ass_item(c, idx, pyitem)\n", sym.id)
		g.wrapper.Outdent()
		g.wrapper.Printf("return c\n")
	case sym.isSlice():
		g.wrapper.Printf("if not isinstance(o, collections.Iterable):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__init__ takes a sequence as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Slice)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("c = _cffi_helper.lib.cgo_func_%[1]s_new()\n", sym.id)
		g.wrapper.Printf("for elt in o:\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(elt)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_append(c, pyitem)\n", sym.id)
		g.wrapper.Outdent()
		g.wrapper.Printf("return c\n")
	case sym.isMap():
		g.wrapper.Printf("if not isinstance(o, collections.Mapping):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise TypeError('%[1]s.__init__ takes a mapping as argument')\n", sym.goname)
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Map)
		esym := g.pkg.syms.symtype(typ.Elem())
		ksym := g.pkg.syms.symtype(typ.Key())
		g.wrapper.Printf("c = _cffi_helper.lib.cgo_func_%[1]s_new()\n", sym.id)
		g.wrapper.Printf("for k, v in o.items():\n")
		g.wrapper.Indent()
		g.wrapper.Printf("pykey = _cffi_helper.cffi_%[1]s(k)\n", ksym.py2c)
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(v)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_set(c, pykey, pyitem)\n", sym.id)
		g.wrapper.Outdent()
		g.wrapper.Printf("return c\n")
	}
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
	g.wrapper.Printf("@staticmethod\n")
	g.wrapper.Printf("def cffi_cgopy_cnv_c2py_%[1]s(c):\n", sym.id)
	g.wrapper.Indent()
	g.wrapper.Printf("o = %[1]s()\n", typename)
	g.wrapper.Printf("o.cgopy = c\n")
	g.wrapper.Printf("return o\n")
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeMethod generates Type methods.
func (g *cffiGen) genTypeMethod(sym *symbol) {
	g.wrapper.Printf("# methods for %s\n", sym.gofmt())
	if sym.isNamed() {
		typ := sym.GoType().(*types.Named)
		for imeth := 0; imeth < typ.NumMethods(); imeth++ {
			m := typ.Method(imeth)
			if !m.Exported() {
				continue
			}
			mname := types.ObjectString(m, nil)
			msym := g.pkg.syms.sym(mname)
			if msym == nil {
				panic(fmt.Errorf(
					"gopy: could not find symbol for [%[1]T] (%#[1]v) (%[2]s)",
					m.Type(),
					m.Name()+" || "+m.FullName(),
				))
			}
			g.genTypeFunc(sym, msym)
		}
	}
	g.wrapper.Printf("\n")
}

// genTypeTPStr generates Type __str__ method.
func (g *cffiGen) genTypeTPRepr(sym *symbol) {
	g.wrapper.Printf("def __repr__(self):\n")
	g.wrapper.Indent()
	g.wrapper.Printf("cret = _cffi_helper.lib.cgo_func_%[1]s_str(self.cgopy)\n", sym.id)
	g.wrapper.Printf("ret = _cffi_helper.cffi_cgopy_cnv_c2py_string(cret)\n")
	g.wrapper.Printf("return ret\n")
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeTPStr generates __str__ method of a type.
func (g *cffiGen) genTypeTPStr(sym *symbol) {
	g.wrapper.Printf("def __str__(self):\n")
	g.wrapper.Indent()
	g.wrapper.Printf("cret = _cffi_helper.lib.cgo_func_%[1]s_str(self.cgopy)\n", sym.id)
	g.wrapper.Printf("ret = _cffi_helper.cffi_cgopy_cnv_c2py_string(cret)\n")
	g.wrapper.Printf("return ret\n")
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

// genTypeSetItem generates __setitem__ of a type.
func (g *cffiGen) genTypeSetItem(sym *symbol) {
	g.wrapper.Printf("def __setitem__(self, idx, item):\n")
	g.wrapper.Indent()
	switch {
	case sym.isArray():
		g.wrapper.Printf("if idx >= len(self):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise IndexError('array index out of range')\n")
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Array)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(item)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_ass_item(self.cgopy, idx, pyitem)\n", sym.id)
	case sym.isSlice():
		g.wrapper.Printf("if idx >= len(self):\n")
		g.wrapper.Indent()
		g.wrapper.Printf("raise IndexError('slice index out of range')\n")
		g.wrapper.Outdent()
		typ := sym.GoType().Underlying().(*types.Slice)
		esym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(item)\n", esym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_ass_item(self.cgopy, idx, pyitem)\n", sym.id)
	case sym.isMap():
		typ := sym.GoType().Underlying().(*types.Map)
		ksym := g.pkg.syms.symtype(typ.Key())
		vsym := g.pkg.syms.symtype(typ.Elem())
		g.wrapper.Printf("pykey = _cffi_helper.cffi_%[1]s(idx)\n", ksym.py2c)
		g.wrapper.Printf("pyitem = _cffi_helper.cffi_%[1]s(item)\n", vsym.py2c)
		g.wrapper.Printf("_cffi_helper.lib.cgo_func_%[1]s_set(self.cgopy, pykey, pyitem)\n", sym.id)
	default:
		panic(fmt.Errorf(
			"gopy: __setitem__ for %s not handled",
			sym.gofmt(),
		))
	}
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}
