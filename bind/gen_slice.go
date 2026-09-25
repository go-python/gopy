// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"go/types"
	"strings"
)

// extTypes = these are types external to any targeted packages
// pyWrapOnly = only generate python wrapper code, not go code
// slob = official slice object -- for package-processed slices -- has methods
func (g *pyGen) genSlice(slc *symbol, extTypes, pyWrapOnly bool, slob *Slice) {
	if slc.isPointer() {
		return // TODO: not sure what to do..
	}
	_, isslc := slc.GoType().Underlying().(*types.Slice)
	_, isary := slc.GoType().Underlying().(*types.Array)
	if !isslc && !isary {
		return
	}

	pkgname := slc.gopkg.Name()

	pysnm := slc.id
	if !strings.Contains(pysnm, "Slice_") {
		pysnm = strings.TrimPrefix(pysnm, pkgname+"_")
	}

	gocl := "go."
	if g.pkg == goPackage {
		gocl = ""
	}

	if !extTypes || pyWrapOnly {
		// TODO: inherit from collections.Iterable too?
		g.pywrap.Printf(`
# Python type for slice %[4]s
class %[2]s(%[5]sGoClass):
	""%[3]q""
`,
			pkgname,
			pysnm,
			slc.doc,
			slc.goname,
			gocl,
		)
		g.pywrap.Indent()
	}

	g.genSliceInit(slc, extTypes, pyWrapOnly, slob)
	if slob != nil {
		g.genSliceMethods(slob)
	}
	if !extTypes || pyWrapOnly {
		g.pywrap.Outdent()
	}
}

func (g *pyGen) genSliceInit(slc *symbol, extTypes, pyWrapOnly bool, slob *Slice) {
	pkgname := slc.gopkg.Name()
	slNm := slc.id
	qNm := g.cfg.Name + "." + slNm // this is only for referring to the _ .go package!
	var esym *symbol
	if typ, ok := slc.GoType().Underlying().(*types.Slice); ok {
		esym = current.symtype(typ.Elem())
	} else if typ, ok := slc.GoType().Underlying().(*types.Array); ok {
		esym = current.symtype(typ.Elem())
	}

	// element access (elem/set/append) below would reference *C.PyObject,
	// which cffi's preamble doesn't declare (see genFuncSig for the same
	// restriction on plain function args/returns); skip the whole wrapper
	// rather than emit code that fails to compile.
	if g.noAPIShim() && esym != nil && esym.cpyname == "PyObject*" && !g.isCFFIComplex(esym) {
		return
	}

	gocl := "go."
	if g.pkg == goPackage {
		gocl = ""
	}

	pysnm := slc.id
	if !strings.Contains(pysnm, "Slice_") {
		pysnm = strings.TrimPrefix(pysnm, pkgname+"_")
	}

	if !extTypes || pyWrapOnly {
		g.pywrap.Printf("def __init__(self, *args, **kwargs):\n")
		g.pywrap.Indent()
		g.pywrap.Printf(`"""
handle=A Go-side object is always initialized with an explicit handle=arg
otherwise parameter is a python list that we copy from
"""
`)
		g.pywrap.Printf("self.index = 0\n")
		g.pywrap.Printf("if len(kwargs) == 1 and 'handle' in kwargs:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = kwargs['handle']\n")
		g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
		g.pywrap.Outdent()
		g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], %sGoClass):\n", gocl)
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = args[0].handle\n")
		g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
		g.pywrap.Outdent()
		if slc.isSlice() {
			g.pywrap.Printf("else:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("self.handle = _%s_CTor()\n", qNm)
			g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
			g.pywrap.Printf("if len(args) > 0:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("if not isinstance(args[0], _collections_abc.Iterable):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("raise TypeError('%s.__init__ takes a sequence as argument')\n", slNm)
			g.pywrap.Outdent()
			g.pywrap.Printf("for elt in args[0]:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("self.append(elt)\n")
			g.pywrap.Outdent()
			g.pywrap.Outdent()
			g.pywrap.Outdent()
		}
		g.pywrap.Outdent()

		g.pywrap.Printf("def __del__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("_%s.DecRef(self.handle)\n", g.pypkgname)
		g.pywrap.Outdent()

		if slob != nil && slob.prots&ProtoStringer != 0 {
			for _, m := range slob.meths {
				if isStringer(m.obj) {
					g.pywrap.Printf("def __str__(self):\n")
					g.pywrap.Indent()
					g.genStringerCall()
					g.pywrap.Outdent()
				}
			}
		} else {
			g.pywrap.Printf("def __str__(self):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("s = '%s.%s len: ' + str(len(self)) + ' handle: ' + str(self.handle) + ' ['\n", pkgname, slNm)
			g.pywrap.Printf("if len(self) < 120:\n")
			g.pywrap.Indent()
			g.pywrap.Println("s += ', '.join(map(str, self)) + ']'")
			g.pywrap.Outdent()
			g.pywrap.Println("return s")
			g.pywrap.Outdent()
		}

		g.pywrap.Printf("def __repr__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return '%s.%s([' + ', '.join(map(str, self)) + '])'\n", pkgname, slNm)
		g.pywrap.Outdent()

		g.pywrap.Printf("def __len__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return _%s_len(self.handle)\n", qNm)
		g.pywrap.Outdent()

		g.pywrap.Printf("def __getitem__(self, key):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if isinstance(key, slice):\n")
		g.pywrap.Indent()
		if slc.isSlice() {
			g.pywrap.Printf("if key.step == None or key.step == 1:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("st = key.start\n")
			g.pywrap.Printf("ed = key.stop\n")
			g.pywrap.Printf("if st == None:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("st = 0\n")
			g.pywrap.Outdent()
			g.pywrap.Printf("if ed == None:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("ed = _%s_len(self.handle)\n", qNm)
			g.pywrap.Outdent()
			g.pywrap.Printf("return %s(handle=_%s_subslice(self.handle, st, ed))\n", pysnm, qNm)
			g.pywrap.Outdent()
		}
		g.pywrap.Printf("return [self[ii] for ii in range(*key.indices(len(self)))]\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("elif isinstance(key, int):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if key < 0:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("key += len(self)\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("if key < 0 or key >= len(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("raise IndexError('slice index out of range')\n")
		g.pywrap.Outdent()
		if esym.hasHandle() {
			g.pywrap.Printf("return %s(handle=_%s_elem(self.handle, key))\n", esym.pyPkgId(slc.gopkg), qNm)
		} else {
			g.pywrap.Printf("return _%s_elem(self.handle, key)\n", qNm)
		}
		g.pywrap.Outdent()
		g.pywrap.Printf("else:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("raise TypeError('slice index invalid type')\n")
		g.pywrap.Outdent()
		g.pywrap.Outdent()

		g.pywrap.Printf("def __setitem__(self, idx, value):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if idx < 0:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("idx += len(self)\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("if idx < len(self):\n")
		g.pywrap.Indent()
		if esym.hasHandle() {
			g.pywrap.Printf("_%s_set(self.handle, idx, value.handle)\n", qNm)
		} else {
			g.pywrap.Printf("_%s_set(self.handle, idx, value)\n", qNm)
		}
		g.pywrap.Println("return")
		g.pywrap.Outdent()
		g.pywrap.Printf("raise IndexError('slice index out of range')\n")
		g.pywrap.Outdent()

		if slc.isSlice() {
			g.pywrap.Printf("def __iadd__(self, value):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("if not isinstance(value, _collections_abc.Iterable):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("raise TypeError('%s.__iadd__ takes a sequence as argument')\n", slNm)
			g.pywrap.Outdent()
			g.pywrap.Printf("for elt in value:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("self.append(elt)\n")
			g.pywrap.Outdent()
			g.pywrap.Printf("return self\n")
			g.pywrap.Outdent()
		}

		g.pywrap.Printf("def __iter__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Println("self.index = 0")
		g.pywrap.Println("return self")
		g.pywrap.Outdent()

		if g.lang == 2 {
			g.pywrap.Printf("def next(self):\n")
		} else {
			g.pywrap.Printf("def __next__(self):\n")
		}
		g.pywrap.Indent()
		g.pywrap.Printf("if self.index < len(self):\n")
		g.pywrap.Indent()
		if esym.hasHandle() {
			g.pywrap.Printf("rv = %s(handle=_%s_elem(self.handle, self.index))\n", esym.pyPkgId(slc.gopkg), qNm)
		} else {
			g.pywrap.Printf("rv = _%s_elem(self.handle, self.index)\n", qNm)
		}
		g.pywrap.Println("self.index = self.index + 1")
		g.pywrap.Println("return rv")
		g.pywrap.Outdent()
		g.pywrap.Printf("raise StopIteration\n")
		g.pywrap.Outdent()

		if slc.isSlice() {
			g.pywrap.Printf("def append(self, value):\n")
			g.pywrap.Indent()
			if esym.hasHandle() {
				g.pywrap.Printf("_%s_append(self.handle, value.handle)\n", qNm)
			} else {
				g.pywrap.Printf("_%s_append(self.handle, value)\n", qNm)
			}
			g.pywrap.Outdent()
			g.pywrap.Printf("def copy(self, src):\n")
			g.pywrap.Indent()
			g.pywrap.Printf(`""" copy emulates the go copy function, copying elements into this list from source list, up to min of size of each list """
`)
			g.pywrap.Printf("mx = min(len(self), len(src))\n")
			g.pywrap.Printf("for i in range(mx):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("self[i] = src[i]\n")
			g.pywrap.Outdent()
			g.pywrap.Outdent()
		}

		if slNm == "Slice_byte" {
			g.pywrap.Printf("@staticmethod\n")
			g.pywrap.Printf("def from_bytes(value):\n")
			g.pywrap.Indent()
			g.pywrap.Printf(`"""Create a Go []byte object from a Python bytes object"""
`)
			g.pywrap.Printf("handle = _%s_from_bytes(value)\n", qNm)
			g.pywrap.Printf("return Slice_byte(handle=handle)\n")
			g.pywrap.Outdent()
			g.pywrap.Printf("def __bytes__(self):\n")
			g.pywrap.Indent()
			g.pywrap.Printf(`"""Convert the slice to a bytes object."""
`)
			g.pywrap.Printf("return _%s_to_bytes(self.handle)\n", qNm)
			g.pywrap.Outdent()
			g.pywrap.Outdent()

		}
	}

	if !extTypes || !pyWrapOnly {
		// go ctor
		ctNm := slNm + "_CTor"
		g.gofile.Printf("\n// --- wrapping slice: %v ---\n", slc.goname)
		g.gofile.Printf("//export %s\n", ctNm)
		g.gofile.Printf("func %s() CGoHandle {\n", ctNm)
		g.gofile.Indent()
		g.gofile.Printf("return CGoHandle(handleFromPtr_%[1]s(&%[2]s{}))\n", slNm, slc.goname)
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s', retval('%s'), [])\n", ctNm, PyHandle)

		g.gofile.Printf("//export %s_len\n", slNm)
		g.gofile.Printf("func %s_len(handle CGoHandle) int {\n", slNm)
		g.gofile.Indent()
		g.gofile.Printf("return len(deptrFromHandle_%s(handle))\n", slNm)
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_len', retval('int'), [param('%s', 'handle')])\n", slNm, PyHandle)

		g.gofile.Printf("//export %s_elem\n", slNm)
		g.gofile.Printf("func %s_elem(handle CGoHandle, _idx int) %s {\n", slNm, g.cgoResult(esym))
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
		// If the go2py starts with handleFromPtr_, use reference &, otherwise just return the value
		val_str := "s[_idx]"
		if strings.HasPrefix(esym.go2py, "handleFromPtr_") {
			val_str = "&(s[_idx])"
		}
		g.gofile.Printf("return %s\n", g.goToCgo(esym, val_str))
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		var caller_owns_ret string
		var transfer_ownership string
		if esym.cpyname == "PyObject*" {
			caller_owns_ret = ", caller_owns_return=True"
			transfer_ownership = ", transfer_ownership=False"
		}
		if esym.cpyname == "char*" {
			g.pybuild.Printf("add_checked_string_function(mod, '%s_elem', retval('%s'), [param('%s', 'handle'), param('int', 'idx')])\n", slNm, esym.cpyname, PyHandle)
		} else {
			g.pybuild.Printf("mod.add_function('%s_elem', retval('%s'%s), [param('%s', 'handle'), param('int', 'idx')])\n", slNm, g.cpyName(esym), caller_owns_ret, PyHandle)
		}

		if slc.isSlice() {
			g.gofile.Printf("//export %s_subslice\n", slNm)
			g.gofile.Printf("func %s_subslice(handle CGoHandle, _st, _ed int) CGoHandle {\n", slNm)
			g.gofile.Indent()
			g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
			g.gofile.Printf("ss := s[_st:_ed]\n")
			g.gofile.Printf("return CGoHandle(handleFromPtr_%s(&ss))\n", slNm)
			g.gofile.Outdent()
			g.gofile.Printf("}\n\n")

			g.pybuild.Printf("mod.add_function('%s_subslice', retval('%s'), [param('%s', 'handle'), param('int', 'st'), param('int', 'ed')])\n", slNm, PyHandle, PyHandle)
		}

		g.gofile.Printf("//export %s_set\n", slNm)
		g.gofile.Printf("func %s_set(handle CGoHandle, _idx int, %s) {\n", slNm, g.cgoParam("_vl", esym))
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
		g.gofile.Printf("s[_idx] = %s\n", g.cgoToGo(esym, "_vl"))
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_set', None, [param('%s', 'handle'), param('int', 'idx'), param('%v', 'value'%s)])\n", slNm, PyHandle, g.cpyName(esym), transfer_ownership)

		if slc.isSlice() {
			g.gofile.Printf("//export %s_append\n", slNm)
			g.gofile.Printf("func %s_append(handle CGoHandle, %s) {\n", slNm, g.cgoParam("_vl", esym))
			g.gofile.Indent()
			g.gofile.Printf("s := ptrFromHandle_%s(handle)\n", slNm)
			g.gofile.Printf("*s = append(*s, %s)\n", g.cgoToGo(esym, "_vl"))
			g.gofile.Outdent()
			g.gofile.Printf("}\n\n")

			g.pybuild.Printf("mod.add_function('%s_append', None, [param('%s', 'handle'), param('%s', 'value'%s)])\n", slNm, PyHandle, g.cpyName(esym), transfer_ownership)
		}

		if slNm == "Slice_byte" {
			if g.noAPIShim() {
				// PyBytes_* is off-limits for cffi (no CPython headers), so these
				// exchange a raw pointer+length instead of a PyObject*; the cffi
				// build script (cffi_build.py) recognizes them by name and writes
				// the bytes<->buffer conversion into the generated python module.
				g.gofile.Printf("//export Slice_byte_from_bytes\n")
				g.gofile.Printf("func Slice_byte_from_bytes(ptr unsafe.Pointer, size C.longlong) CGoHandle {\n")
				g.gofile.Indent()
				g.gofile.Printf("data := make([]byte, size)\n")
				g.gofile.Printf("if size > 0 {\n")
				g.gofile.Indent()
				g.gofile.Printf("tmp := unsafe.Slice((*byte)(ptr), size)\n")
				g.gofile.Printf("copy(data, tmp)\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n")
				g.gofile.Printf("return handleFromPtr_Slice_byte(&data)\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n\n")

				g.gofile.Printf("//export Slice_byte_to_bytes_len\n")
				g.gofile.Printf("func Slice_byte_to_bytes_len(handle CGoHandle) C.longlong {\n")
				g.gofile.Indent()
				g.gofile.Printf("s := deptrFromHandle_Slice_byte(handle)\n")
				g.gofile.Printf("return C.longlong(len(s))\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n\n")

				// Returning &s[0] directly would hand cgo a pointer into the Go
				// heap, which cgo's pointer checks reject once it crosses back
				// to the caller; copy into a C-owned buffer instead, freed by
				// the python side (Slice_byte_free_ptr) once it has read it.
				g.gofile.Printf("//export Slice_byte_to_bytes_ptr\n")
				g.gofile.Printf("func Slice_byte_to_bytes_ptr(handle CGoHandle) unsafe.Pointer {\n")
				g.gofile.Indent()
				g.gofile.Printf("s := deptrFromHandle_Slice_byte(handle)\n")
				g.gofile.Printf("n := len(s)\n")
				g.gofile.Printf("if n == 0 {\n")
				g.gofile.Indent()
				g.gofile.Printf("return nil\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n")
				g.gofile.Printf("buf := C.malloc(C.size_t(n))\n")
				g.gofile.Printf("copy(unsafe.Slice((*byte)(buf), n), s)\n")
				g.gofile.Printf("return buf\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n\n")

				g.gofile.Printf("//export Slice_byte_free_ptr\n")
				g.gofile.Printf("func Slice_byte_free_ptr(ptr unsafe.Pointer) {\n")
				g.gofile.Indent()
				g.gofile.Printf("C.free(ptr)\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n\n")
			} else {
				g.gofile.Printf("//export Slice_byte_from_bytes\n")
				g.gofile.Printf("func Slice_byte_from_bytes(o *C.PyObject) CGoHandle {\n")
				g.gofile.Indent()
				g.gofile.Printf("size := C.PyBytes_Size(o)\n")
				g.gofile.Printf("ptr := unsafe.Pointer(C.PyBytes_AsString(o))\n")
				g.gofile.Printf("data := make([]byte, size)\n")
				g.gofile.Printf("tmp := unsafe.Slice((*byte)(ptr), size)\n")
				g.gofile.Printf("copy(data, tmp)\n")
				g.gofile.Printf("return handleFromPtr_Slice_byte(&data)\n")
				g.gofile.Outdent()
				g.gofile.Printf("}\n\n")

				g.gofile.Printf("//export Slice_byte_to_bytes\n")
				g.gofile.Printf("func Slice_byte_to_bytes(handle CGoHandle) *C.PyObject {\n")
				g.gofile.Indent()
				g.gofile.Printf("s := deptrFromHandle_Slice_byte(handle)\n")
				g.gofile.Printf("ptr := unsafe.Pointer(&s[0])\n")
				g.gofile.Printf("size := len(s)\n")
				if WindowsOS {
					g.gofile.Printf("return C.PyBytes_FromStringAndSize((*C.char)(ptr), C.longlong(size))\n")
				} else {
					g.gofile.Printf("return C.PyBytes_FromStringAndSize((*C.char)(ptr), C.long(size))\n")
				}
				g.gofile.Outdent()
				g.gofile.Printf("}\n\n")

				g.pybuild.Printf("mod.add_function('Slice_byte_from_bytes', retval('%s'%s), [param('PyObject*', 'o', transfer_ownership=False)])\n", PyHandle, caller_owns_ret)
				g.pybuild.Printf("mod.add_function('Slice_byte_to_bytes', retval('PyObject*', caller_owns_return=True), [param('%s', 'handle')])\n", PyHandle)
			}
		}
	}
}

func (g *pyGen) genSliceMethods(s *Slice) {
	for _, m := range s.meths {
		g.genMethod(s.sym, m)
	}
}
