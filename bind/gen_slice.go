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
		return // todo: not sure what to do..
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
		// todo: inherit from collections.Iterable too?
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
	pkgname := g.outname
	slNm := slc.id
	qNm := pkgname + "." + slNm
	var esym *symbol
	if typ, ok := slc.GoType().Underlying().(*types.Slice); ok {
		esym = current.symtype(typ.Elem())
	} else if typ, ok := slc.GoType().Underlying().(*types.Array); ok {
		esym = current.symtype(typ.Elem())
	}

	gocl := "go."
	if g.pkg == goPackage {
		gocl = ""
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
		g.pywrap.Outdent()
		g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], %sGoClass):\n", gocl)
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = args[0].handle\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], int):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = args[0]\n")
		g.pywrap.Outdent()
		if slc.isSlice() {
			g.pywrap.Printf("else:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("self.handle = _%s_CTor()\n", qNm)
			g.pywrap.Printf("if len(args) > 0:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("if not isinstance(args[0], collections.Iterable):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("raise TypeError('%s.__init__ takes a sequence as argument')\n", slNm)
			g.pywrap.Outdent()
			g.pywrap.Printf("for elt in args[0]:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("_%s_append(self.handle, elt)\n", qNm)
			g.pywrap.Outdent()
			g.pywrap.Outdent()
			g.pywrap.Outdent()
		}
		g.pywrap.Outdent()

		if slob != nil && slob.prots&ProtoStringer != 0 {
			for _, m := range slob.meths {
				if isStringer(m.obj) {
					g.pywrap.Printf("def __str__(self):\n")
					g.pywrap.Indent()
					g.pywrap.Printf("return self.String()\n")
					g.pywrap.Outdent()
					g.pywrap.Printf("\n")
				}
			}
		}

		g.pywrap.Printf("def __len__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return _%s_len(self.handle)\n", qNm)
		g.pywrap.Outdent()

		g.pywrap.Printf("def __getitem__(self, idx):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if idx >= len(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("raise IndexError('slice index out of range')\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("return _%s_elem(self.handle, idx)\n", qNm)
		g.pywrap.Outdent()

		if slc.isSlice() {
			g.pywrap.Printf("def __setitem__(self, idx, value):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("if idx >= len(self):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("raise IndexError('slice index out of range')\n")
			g.pywrap.Outdent()
			if esym.hasHandle() {
				g.pywrap.Printf("_%s_set(self.handle, idx, value.handle)\n", qNm)
			} else {
				g.pywrap.Printf("_%s_set(self.handle, idx, value)\n", qNm)
			}
			g.pywrap.Outdent()

			g.pywrap.Printf("def __iadd__(self, value):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("if not isinstance(value, collections.Iterable):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("raise TypeError('%s.__iadd__ takes a sequence as argument')\n", slNm)
			g.pywrap.Outdent()
			g.pywrap.Printf("for elt in value:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("_%s_append(self.handle, elt)\n", qNm)
			g.pywrap.Outdent()
			g.pywrap.Outdent()
		}

		g.pywrap.Printf("def __iter__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return self\n")
		g.pywrap.Outdent()

		g.pywrap.Printf("def __next__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if self.index >= len(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("raise StopIteration\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("self.index = self.index + 1\n")
		g.pywrap.Printf("return _%s_elem(self.handle, self.index)\n", qNm)
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

		// g.pywrap.Printf("	def __repr__(self):
		//        cret = _cffi_helper.lib.cgo_func_0x3243646956_str(self.cgopy)
		//        ret = _cffi_helper.cffi_cgopy_cnv_c2py_string(cret)
		//        return ret

	}

	if !extTypes || !pyWrapOnly {
		// go ctor
		ctNm := slNm + "_CTor"
		g.gofile.Printf("\n// --- wrapping slice: %v ---\n", slc.goname)
		g.gofile.Printf("//export %s\n", ctNm)
		g.gofile.Printf("func %s() CGoHandle {\n", ctNm)
		g.gofile.Indent()
		g.gofile.Printf("return CGoHandle(handleFmPtr_%[1]s(&%[2]s{}))\n", slNm, slc.gofmt())
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s', retval('%s'), [])\n", ctNm, PyHandle)

		g.gofile.Printf("//export %s_len\n", slNm)
		g.gofile.Printf("func %s_len(handle CGoHandle) int {\n", slNm)
		g.gofile.Indent()
		g.gofile.Printf("return len(*ptrFmHandle_%s(handle))\n", slNm)
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_len', retval('int'), [param('%s', 'handle')])\n", slNm, PyHandle)

		g.gofile.Printf("//export %s_elem\n", slNm)
		g.gofile.Printf("func %s_elem(handle CGoHandle, _idx int) %s {\n", slNm, esym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := *ptrFmHandle_%s(handle)\n", slNm)
		if esym.go2py != "" {
			g.gofile.Printf("return %s(s[_idx])%s\n", esym.go2py, esym.go2pyParenEx)
		} else {
			g.gofile.Printf("return s[_idx]\n")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_elem', retval('%s'), [param('%s', 'handle'), param('int', 'idx')])\n", slNm, esym.cpyname, PyHandle)

		g.gofile.Printf("//export %s_set\n", slNm)
		g.gofile.Printf("func %s_set(handle CGoHandle, _idx int, _vl %s) {\n", slNm, esym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := *ptrFmHandle_%s(handle)\n", slNm)
		if esym.py2go != "" {
			g.gofile.Printf("s[_idx] = %s(_vl)%s\n", esym.py2go, esym.py2goParenEx)
		} else {
			g.gofile.Printf("s[_idx] = _vl\n")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_set', None, [param('%s', 'handle'), param('int', 'idx'), param('%v', 'value')])\n", slNm, PyHandle, esym.cpyname)

		if slc.isSlice() {
			g.gofile.Printf("//export %s_append\n", slNm)
			g.gofile.Printf("func %s_append(handle CGoHandle, _vl %s) {\n", slNm, esym.cgoname)
			g.gofile.Indent()
			g.gofile.Printf("s := ptrFmHandle_%s(handle)\n", slNm)
			if esym.py2go != "" {
				g.gofile.Printf("*s = append(*s, %s(_vl)%s)\n", esym.py2go, esym.py2goParenEx)
			} else {
				g.gofile.Printf("*s = append(*s, _vl)\n")
			}
			g.gofile.Outdent()
			g.gofile.Printf("}\n\n")

			g.pybuild.Printf("mod.add_function('%s_append', None, [param('%s', 'handle'), param('%s', 'value')])\n", slNm, PyHandle, esym.cpyname)
		}
	}
}

func (g *pyGen) genSliceMethods(s *Slice) {
	for _, m := range s.meths {
		g.genMethod(s.sym, m)
	}
}