// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/token"
)

const (
	builderPreamble = `import os

from sys import argv
from cffi import FFI

os.chdir(os.path.dirname(os.path.abspath(__file__)))
ffibuilder = FFI()
ffibuilder.set_source(
    'gopy.%[1]s',
    None,
    extra_objects=["gopy/_%[1]s.so"],
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

def touch(fname):
    try:
        os.utime(fname, None)
    except OSError:
        open(fname, 'a').close()

if __name__ == "__main__":
    ffibuilder.compile()
    # touch __init__.py
    touch('gopy/__init__.py')
    # Remove itself after compile.
    os.remove(argv[0])
`

	cffiPreamble = `"""%[1]s"""
from __future__ import absolute_import

import os
from gopy.%[2]s import ffi

here = os.path.dirname(os.path.abspath(__file__))
lib = ffi.dlopen(os.path.join(here, "gopy/_%[2]s.so"))

# python <--> cffi helpers.
def cffi_cnv_py2c_string(o):
    s = ffi.new("char[]", o)
    return lib._cgopy_GoString(s)

def cffi_cnv_py2c_int(o):
    return ffi.cast('int', o)

def cffi_cnv_py2c_float32(o):
    return ffi.cast('float', o)

def cffi_cnv_py2c_float64(o):
    return ffi.cast('double', o)

def cffi_cnv_c2py_string(c):
    s = lib._cgopy_CString(c)
    return ffi.string(s)

def cffi_cnv_c2py_int(c):
    return int(c)

def cffi_cnv_c2py_float32(c):
    return float(c)

def cffi_cnv_c2py_float64(c):
    return float(c)

# make sure Cgo is loaded and initialized
lib.cgo_pkg_%[2]s_init()

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
	package_doc := g.pkg.doc.Doc
	g.wrapper.Printf(cffiPreamble, package_doc, n)
}
