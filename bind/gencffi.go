// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/token"
)

const (
	/*
	   FIXME(corona10): ffibuilder.cdef should be written this way.
	   ffibuilder.cdef("""
	   	//header exported from 'go tool cgo'
	   	#include "%[3]s.h"
	   	""")

	   * discuss: https://github.com/go-python/gopy/pull/93#discussion_r119652220
	*/
	builderPreamble = `import os

from sys import argv
from cffi import FFI

ffibuilder = FFI()
ffibuilder.set_source(
    '%[1]s',
    None,
    extra_objects=["_%[1]s.so"],
)

ffibuilder.cdef("""
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

extern GoString _cgopy_GoString(char* p0);
extern char* _cgopy_CString(GoString p0);
extern GoUint8 _cgopy_ErrorIsNil(GoInterface p0);
extern char* _cgopy_ErrorString(GoInterface p0);
extern void cgopy_incref(void* p0);
extern void cgopy_decref(void* p0);

extern void cgo_pkg_%[1]s_init();

`
	builderPreambleEnd = `""")

if __name__ == "__main__":
    ffibuilder.compile()
    # Remove itself after compile.
    os.remove(argv[0])
`

	cffiPreamble = `
__doc__="""%[1]s"""

import os

# python <--> cffi helper.
class _cffi_helper(object):

    here = os.path.dirname(os.path.abspath(__file__))
    lib = ffi.dlopen(os.path.join(here, "_%[2]s.so"))

    @staticmethod
    def cffi_cnv_py2c_string(o):
        s = ffi.new("char[]", o)
        return _cffi_helper.lib._cgopy_GoString(s)

    @staticmethod
    def cffi_cnv_py2c_int(o):
        return ffi.cast('int', o)

    @staticmethod
    def cffi_cnv_py2c_float32(o):
        return ffi.cast('float', o)

    @staticmethod
    def cffi_cnv_py2c_float64(o):
        return ffi.cast('double', o)

    @staticmethod
    def cffi_cnv_c2py_string(c):
        s = _cffi_helper.lib._cgopy_CString(c)
        return ffi.string(s)

    @staticmethod
    def cffi_cnv_c2py_int(c):
        return int(c)

    @staticmethod
    def cffi_cnv_c2py_float32(c):
        return float(c)

    @staticmethod
    def cffi_cnv_c2py_float64(c):
        return float(c)

# make sure Cgo is loaded and initialized
_cffi_helper.lib.cgo_pkg_%[2]s_init()
`
)

type cffiGen struct {
	builder *printer
	wrapper *printer

	fset *token.FileSet
	pkg  *Package
	err  ErrorList

	lang int // c-python api version (2,3)
}

func (g *cffiGen) gen() error {
	// Write preamble for CFFI builder.py
	g.genBuilderPreamble()
	// Write preamble for CFFI library wrapper.
	g.genCffiPreamble()
	for _, f := range g.pkg.funcs {
		g.genFunc(f)
		g.genCdef(f)
	}

	// Finalizing preamble for CFFI builder.py
	g.genBuilderPreambleEnd()
	return nil
}

func (g *cffiGen) genBuilderPreamble() {
	n := g.pkg.pkg.Name()
	g.builder.Printf(builderPreamble, n)
}

func (g *cffiGen) genBuilderPreambleEnd() {
	g.builder.Printf(builderPreambleEnd)
}

func (g *cffiGen) genCffiPreamble() {
	n := g.pkg.pkg.Name()
	pkgDoc := g.pkg.doc.Doc
	g.wrapper.Printf(cffiPreamble, pkgDoc, n)
}
