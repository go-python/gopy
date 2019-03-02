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
// mpob = Map object -- for methods
func (g *pyGen) genMap(slc *symbol, extTypes, pyWrapOnly bool, mpob *Map) {
	if slc.isPointer() {
		return // todo: not sure what to do..
	}
	_, ok := slc.GoType().Underlying().(*types.Map)
	if !ok {
		return
	}

	pkgname := slc.gopkg.Name()

	// todo: maybe check for named type here or something?
	pysnm := slc.id
	if !strings.Contains(pysnm, "Map_") {
		pysnm = strings.TrimPrefix(pysnm, pkgname+"_")
	}

	gocl := "go."
	if g.pkg == goPackage {
		gocl = ""
	}

	if !extTypes || pyWrapOnly {
		// todo: inherit from collections.Iterable too?
		g.pywrap.Printf(`
# Python type for map %[4]s
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

	g.genMapInit(slc, extTypes, pyWrapOnly)
	if mpob != nil {
		g.genMapMethods(mpob)
	}
	if !extTypes || pyWrapOnly {
		g.pywrap.Outdent()
	}
}

func (g *pyGen) genMapInit(slc *symbol, extTypes, pyWrapOnly bool) {
	pkgname := g.outname
	slNm := slc.id
	qNm := pkgname + "." + slNm
	typ := slc.GoType().Underlying().(*types.Map)
	esym := current.symtype(typ.Elem())
	ksym := current.symtype(typ.Key())

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
		g.pywrap.Printf("else:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = _%s_CTor()\n", qNm)
		g.pywrap.Printf("if len(args) > 0:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if not isinstance(args[0], collections.Mapping):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("raise TypeError('%s.__init__ takes a mapping as argument')\n", slNm)
		g.pywrap.Outdent()
		g.pywrap.Printf("for k, v in args[0].items():\n")
		g.pywrap.Indent()
		g.pywrap.Printf("_%s_set(self.handle, k, v)\n", qNm)
		g.pywrap.Outdent()
		g.pywrap.Outdent()
		g.pywrap.Outdent()
		g.pywrap.Outdent()

		// for _, m := range slc.meths {
		// 	if m.GoName() == "String" {
		// 		g.pywrap.Printf("def __str__(self):\n")
		// 		g.pywrap.Indent()
		// 		g.pywrap.Printf("return self.String()\n")
		// 		g.pywrap.Outdent()
		// 		g.pywrap.Printf("\n")
		// 	}
		// }
		g.pywrap.Printf("def __len__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return _%s_len(self.handle)\n", qNm)
		g.pywrap.Outdent()

		g.pywrap.Printf("def __getitem__(self, key):\n")
		g.pywrap.Indent()
		if ksym.hasHandle() {
			g.pywrap.Printf("return _%s_elem(self.handle, key.handle)\n", qNm)
		} else {
			g.pywrap.Printf("return _%s_elem(self.handle, key)\n", qNm)
		}
		g.pywrap.Outdent()

		g.pywrap.Printf("def __setitem__(self, key, value):\n")
		g.pywrap.Indent()
		if esym.hasHandle() {
			if ksym.hasHandle() {
				g.pywrap.Printf("_%s_set(self.handle, key.handle, value.handle)\n", qNm)
			} else {
				g.pywrap.Printf("_%s_set(self.handle, key, value.handle)\n", qNm)
			}
		} else {
			if ksym.hasHandle() {
				g.pywrap.Printf("_%s_set(self.handle, key.handle, value)\n", qNm)
			} else {
				g.pywrap.Printf("_%s_set(self.handle, key, value)\n", qNm)
			}
		}
		g.pywrap.Outdent()

		// g.pywrap.Printf("def __iter__(self):\n")
		// g.pywrap.Indent()
		// g.pywrap.Printf("return self\n")
		// g.pywrap.Outdent()
		//
		// g.pywrap.Printf("def __next__(self):\n")
		// g.pywrap.Indent()
		// g.pywrap.Printf("if self.index >= len(self):\n")
		// g.pywrap.Indent()
		// g.pywrap.Printf("raise StopIteration\n")
		// g.pywrap.Outdent()
		// g.pywrap.Printf("self.index = self.index + 1\n")
		// g.pywrap.Printf("return _%s_elem(self.handle, self.index)\n", qNm)
		// g.pywrap.Outdent()
	}

	if !extTypes || !pyWrapOnly {
		// go ctor
		ctNm := slNm + "_CTor"
		g.gofile.Printf("\n// --- wrapping map: %v ---\n", slc.goname)
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
		g.gofile.Printf("func %s_elem(handle CGoHandle, _ky %s) %s {\n", slNm, ksym.cgoname, esym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := *ptrFmHandle_%s(handle)\n", slNm)
		if esym.go2py != "" {
			if ksym.py2go != "" {
				g.gofile.Printf("return %s(s[%s(_ky)%s])%s\n", esym.go2py, ksym.py2go, ksym.py2goParenEx, esym.go2pyParenEx)
			} else {
				g.gofile.Printf("return %s(s[_ky])%s\n", esym.go2py, esym.go2pyParenEx)
			}
		} else {
			if ksym.py2go != "" {
				g.gofile.Printf("return s[%s(_ky)%s]\n", ksym.py2go, ksym.py2goParenEx)
			} else {
				g.gofile.Printf("return s[_ky]\n")
			}
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_elem', retval('%s'), [param('%s', 'handle'), param('%s', '_ky')])\n", slNm, esym.cpyname, PyHandle, ksym.cpyname)

		g.gofile.Printf("//export %s_set\n", slNm)
		g.gofile.Printf("func %s_set(handle CGoHandle, _ky %s, _vl %s) {\n", slNm, ksym.cgoname, esym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := *ptrFmHandle_%s(handle)\n", slNm)
		if ksym.py2go != "" {
			g.gofile.Printf("s[%s(_ky)%s] = ", ksym.py2go, ksym.py2goParenEx)
		} else {
			g.gofile.Printf("s[_ky] = ")
		}
		if esym.py2go != "" {
			g.gofile.Printf("%s(_vl)%s\n", esym.py2go, esym.py2goParenEx)
		} else {
			g.gofile.Printf("_vl\n")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_set', None, [param('%s', 'handle'), param('%s', 'key'), param('%s', 'value')])\n", slNm, PyHandle, ksym.cpyname, esym.cpyname)
	}
}

func (g *pyGen) genMapMethods(s *Map) {
	for _, m := range s.meths {
		g.genMethod(s.sym, m)
	}
}
