// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/token"
)

const (
	// FIXME(corona10): ffibuilder.cdef should be written this way.
	// ffi.cdef("""
	//       //header exported from 'go tool cgo'
	//      #include "%[3]s.h"
	//       """)
	// discuss: https://github.com/go-python/gopy/pull/93#discussion_r119652220
	cffiPreamble = `"""%[1]s"""
import collections
import os
import sys
import cffi as _cffi_backend

_PY3 = sys.version_info[0] == 3

ffi = _cffi_backend.FFI()
ffi.cdef("""
typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef size_t GoUintptr;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef float GoFloat32;
typedef double GoFloat64;
typedef struct { const char *p; GoInt n; } GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;
typedef struct { GoFloat32 real; GoFloat32 imag; } GoComplex64;
typedef struct { GoFloat64 real; GoFloat64 imag; } GoComplex128;

extern GoComplex64 _cgopy_GoComplex64(GoFloat32 p0, GoFloat32 p1);
extern GoComplex128 _cgopy_GoComplex128(GoFloat64 p0, GoFloat64 p1);
extern GoString _cgopy_GoString(char* p0);
extern char* _cgopy_CString(GoString p0);
extern void _cgopy_FreeCString(char* p0);
extern GoUint8 _cgopy_ErrorIsNil(GoInterface p0);
extern char* _cgopy_ErrorString(GoInterface p0);
extern void cgopy_incref(void* p0);
extern void cgopy_decref(void* p0);

extern void cgo_pkg_%[2]s_init();

`
	cffiHelperPreamble = `""")

# python <--> cffi helper.
class _cffi_helper(object):

    here = os.path.dirname(os.path.abspath(__file__))
    lib = ffi.dlopen(os.path.join(here, "_%[1]s.so"))

    @staticmethod
    def cffi_cgopy_cnv_py2c_bool(o):
        return ffi.cast('_Bool', o)

    @staticmethod
    def cffi_cgopy_cnv_py2c_complex64(o):
        real = o.real
        imag = o.imag
        complex64 = _cffi_helper.lib._cgopy_GoComplex64(real, imag)
        return complex64

    @staticmethod
    def cffi_cgopy_cnv_py2c_complex128(o):
        real = o.real
        imag = o.imag
        complex128 = _cffi_helper.lib._cgopy_GoComplex128(real, imag)
        return complex128

    @staticmethod
    def cffi_cgopy_cnv_py2c_string(o):
        if _PY3:
            o = o.encode('ascii')
        s = ffi.new("char[]", o)
        return _cffi_helper.lib._cgopy_GoString(s)

    @staticmethod
    def cffi_cgopy_cnv_py2c_int(o):
        return ffi.cast('int', o)

    @staticmethod
    def cffi_cgopy_cnv_py2c_float32(o):
        return ffi.cast('float', o)

    @staticmethod
    def cffi_cgopy_cnv_py2c_float64(o):
        return ffi.cast('double', o)

    @staticmethod
    def cffi_cgopy_cnv_py2c_uint(o):
        return ffi.cast('uint', o)

    @staticmethod
    def cffi_cgopy_cnv_c2py_bool(c):
        return bool(c)

    @staticmethod
    def cffi_cgopy_cnv_c2py_complex64(c):
         return complex(c.real, c.imag)

    @staticmethod
    def cffi_cgopy_cnv_c2py_complex128(c):
         return complex(c.real, c.imag)

    @staticmethod
    def cffi_cgopy_cnv_c2py_string(c):
        s = _cffi_helper.lib._cgopy_CString(c)
        pystr = ffi.string(s)
        _cffi_helper.lib._cgopy_FreeCString(s)
        if _PY3:
            pystr = pystr.decode('utf8')
        return pystr

    @staticmethod
    def cffi_cgopy_cnv_c2py_errstring(c):
        s = _cffi_helper.lib._cgopy_ErrorString(c)
        pystr = ffi.string(s)
        _cffi_helper.lib._cgopy_FreeCString(s)
        if _PY3:
            pystr = pystr.decode('utf8')
        return pystr

    @staticmethod
    def cffi_cgopy_cnv_c2py_int(c):
        return int(c)

    @staticmethod
    def cffi_cgopy_cnv_c2py_float32(c):
        return float(c)

    @staticmethod
    def cffi_cgopy_cnv_c2py_float64(c):
        return float(c)

    @staticmethod
    def cffi_cgopy_cnv_c2py_uint(c):
        return int(c)

`
)

type cffiGen struct {
	wrapper *printer

	fset *token.FileSet
	pkg  *Package
	err  ErrorList

	lang int // c-python api version (2,3)
}

func (g *cffiGen) gen() error {
	// Write preamble for CFFI library wrapper.
	g.genCffiPreamble()
	g.genCffiCdef()
	g.genWrappedPy()
	return nil
}

func (g *cffiGen) genCffiPreamble() {
	n := g.pkg.pkg.Name()
	pkgDoc := g.pkg.doc.Doc
	g.wrapper.Printf(cffiPreamble, pkgDoc, n)
}

func (g *cffiGen) genCffiCdef() {

	// first, process slices, arrays
	{
		names := g.pkg.syms.names()
		for _, n := range names {
			sym := g.pkg.syms.sym(n)
			if !sym.isType() {
				continue
			}
			g.genCdefType(sym)
		}
	}

	// Register struct type defintions
	for _, s := range g.pkg.structs {
		g.genCdefStruct(s)
	}

	for _, s := range g.pkg.structs {
		for _, ctor := range s.ctors {
			g.genCdefFunc(ctor)
		}
	}

	for _, s := range g.pkg.structs {
		for _, m := range s.meths {
			g.genCdefMethod(m)
		}

		typ := s.Struct()
		for i := 0; i < typ.NumFields(); i++ {
			f := typ.Field(i)
			if !f.Exported() {
				continue
			}
			g.genCdefStructMemberGetter(s, i, f)
			g.genCdefStructMemberSetter(s, i, f)
		}

		g.genCdefStructTPStr(s)
	}

	for _, f := range g.pkg.funcs {
		g.genCdefFunc(f)
	}

	for _, c := range g.pkg.consts {
		g.genCdefConst(c)
	}

	for _, v := range g.pkg.vars {
		g.genCdefVar(v)
	}
}

func (g *cffiGen) genWrappedPy() {
	n := g.pkg.pkg.Name()
	g.wrapper.Printf(cffiHelperPreamble, n)
	g.wrapper.Indent()

	// first, process slices, arrays
	names := g.pkg.syms.names()
	for _, n := range names {
		sym := g.pkg.syms.sym(n)
		if !sym.isType() {
			continue
		}
		g.genTypeConverter(sym)
	}

	for _, s := range g.pkg.structs {
		g.genStructConversion(s)
	}
	g.wrapper.Outdent()

	// After generating all of the stuff for the preamble (includes struct, interface.. etc)
	// then call a function which checks Cgo is successfully loaded and initialized.
	g.wrapper.Printf("# make sure Cgo is loaded and initialized\n")
	g.wrapper.Printf("_cffi_helper.lib.cgo_pkg_%[1]s_init()\n", n)

	for _, n := range names {
		sym := g.pkg.syms.sym(n)
		if !sym.isType() {
			continue
		}
		g.genType(sym)
	}

	for _, s := range g.pkg.structs {
		g.genStruct(s)
	}

	for _, s := range g.pkg.structs {
		for _, ctor := range s.ctors {
			g.genFunc(ctor)
		}
	}

	for _, f := range g.pkg.funcs {
		g.genFunc(f)
	}

	for _, c := range g.pkg.consts {
		g.genConst(c)
	}

	for _, v := range g.pkg.vars {
		g.genVar(v)
	}
}

func (g *cffiGen) genConst(c Const) {
	g.genGetFunc(c.f)
}

func (g *cffiGen) genVar(v Var) {
	id := g.pkg.Name() + "_" + v.Name()
	get := "returns " + g.pkg.Name() + "." + v.Name()
	set := "sets " + g.pkg.Name() + "." + v.Name()
	if v.doc != "" {
		// if the Go variable had some documentation attached,
		// put it there as well.
		get += "\n\n" + v.doc
		set += "\n\n" + v.doc
	}
	doc := v.doc
	{
		res := []*Var{newVar(g.pkg, v.GoType(), "ret", v.Name(), doc)}
		sig := newSignature(g.pkg, nil, nil, res)
		fget := Func{
			pkg:  g.pkg,
			sig:  sig,
			typ:  nil,
			name: v.Name(),
			id:   id + "_get",
			doc:  get,
			ret:  v.GoType(),
			err:  false,
		}
		g.genGetFunc(fget)
	}
	{
		params := []*Var{newVar(g.pkg, v.GoType(), "arg", v.Name(), doc)}
		sig := newSignature(g.pkg, nil, params, nil)
		fset := Func{
			pkg:  g.pkg,
			sig:  sig,
			typ:  nil,
			name: v.Name(),
			id:   id + "_set",
			doc:  set,
			ret:  nil,
			err:  false,
		}
		g.genSetFunc(fset)
	}
}

func (g *cffiGen) genCdefConst(c Const) {
	g.genCdefFunc(c.f)
}

func (g *cffiGen) genCdefVar(v Var) {
	id := g.pkg.Name() + "_" + v.Name()
	doc := v.doc
	{
		res := []*Var{newVar(g.pkg, v.GoType(), "ret", v.Name(), doc)}
		sig := newSignature(g.pkg, nil, nil, res)
		fget := Func{
			pkg:  g.pkg,
			sig:  sig,
			typ:  nil,
			name: v.Name(),
			id:   id + "_get",
			doc:  "returns " + g.pkg.Name() + "." + v.Name(),
			ret:  v.GoType(),
			err:  false,
		}
		g.genCdefFunc(fget)
	}
	{
		params := []*Var{newVar(g.pkg, v.GoType(), "arg", v.Name(), doc)}
		sig := newSignature(g.pkg, nil, params, nil)
		fset := Func{
			pkg:  g.pkg,
			sig:  sig,
			typ:  nil,
			name: v.Name(),
			id:   id + "_set",
			doc:  "sets " + g.pkg.Name() + "." + v.Name(),
			ret:  nil,
			err:  false,
		}
		g.genCdefFunc(fset)
	}
}
