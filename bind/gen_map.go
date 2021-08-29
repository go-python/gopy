// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
	"strings"
)

// extTypes = these are types external to any targeted packages
// pyWrapOnly = only generate python wrapper code, not go code
// mpob = Map object -- for methods
func (g *pyGen) genMap(slc *symbol, extTypes, pyWrapOnly bool, mpob *Map) {
	if slc.isPointer() {
		return // TODO: not sure what to do..
	}
	_, ok := slc.GoType().Underlying().(*types.Map)
	if !ok {
		return
	}

	pkgname := slc.gopkg.Name()

	// TODO: maybe check for named type here or something?
	pysnm := slc.id
	if !strings.Contains(pysnm, "Map_") {
		pysnm = strings.TrimPrefix(pysnm, pkgname+"_")
	}

	gocl := "go."
	if g.pkg == goPackage {
		gocl = ""
	}

	if !extTypes || pyWrapOnly {
		// TODO: inherit from collections.Iterable too?
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

	g.genMapInit(slc, extTypes, pyWrapOnly, mpob)
	if mpob != nil {
		g.genMapMethods(mpob)
	}
	if !extTypes || pyWrapOnly {
		g.pywrap.Outdent()
	}
}

func (g *pyGen) genMapInit(slc *symbol, extTypes, pyWrapOnly bool, mpob *Map) {
	pkgname := g.cfg.Name
	slNm := slc.id
	qNm := pkgname + "." + slNm
	typ := slc.GoType().Underlying().(*types.Map)
	esym := current.symtype(typ.Elem())
	ksym := current.symtype(typ.Key())

	// key slice type and name
	keyslt := types.NewSlice(typ.Key())
	keyslsym := current.symtype(keyslt)
	if keyslsym == nil {
		fmt.Printf("nil key slice type!: %s map: %s\n", current.fullTypeString(keyslt), slc.goname)
		return
	}
	keyslnm := ""
	if g.pkg == nil {
		keyslnm = keyslsym.id
	} else {
		keyslnm = keyslsym.pyPkgId(g.pkg.pkg)
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
		g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
		g.pywrap.Outdent()
		g.pywrap.Printf("elif len(args) == 1 and isinstance(args[0], %sGoClass):\n", gocl)
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = args[0].handle\n")
		g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
		g.pywrap.Outdent()
		g.pywrap.Printf("else:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("self.handle = _%s_CTor()\n", qNm)
		g.pywrap.Printf("_%s.IncRef(self.handle)\n", g.pypkgname)
		g.pywrap.Printf("if len(args) > 0:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("if not isinstance(args[0], _collections_abc.Mapping):\n")
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

		g.pywrap.Printf("def __del__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("_%s.DecRef(self.handle)\n", g.pypkgname)
		g.pywrap.Outdent()

		if mpob != nil && mpob.prots&ProtoStringer != 0 {
			for _, m := range mpob.meths {
				if isStringer(m.obj) {
					g.pywrap.Printf("def __str__(self):\n")
					g.pywrap.Indent()
					g.pywrap.Printf("return self.String()\n")
					g.pywrap.Outdent()
					g.pywrap.Printf("\n")
				}
			}
		} else {
			g.pywrap.Printf("def __str__(self):\n")
			g.pywrap.Indent()
			g.pywrap.Printf("s = '%s.%s len: ' + str(len(self)) + ' handle: ' + str(self.handle) + ' {'\n", pkgname, slNm)
			g.pywrap.Printf("if len(self) < 120:\n")
			g.pywrap.Indent()
			g.pywrap.Printf("for k, v in self.items():\n")
			g.pywrap.Indent()
			g.pywrap.Printf("s += str(k) + '=' + str(v) + ', '\n")
			g.pywrap.Outdent()
			g.pywrap.Outdent()
			g.pywrap.Println("return s + '}'")
			g.pywrap.Outdent()
		}

		g.pywrap.Printf("def __repr__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("s = '%s.%s({'\n", pkgname, slNm)
		g.pywrap.Printf("for k, v in self.items():\n")
		g.pywrap.Indent()
		g.pywrap.Printf("s += str(k) + '=' + str(v) + ', '\n")
		g.pywrap.Outdent()
		g.pywrap.Println("return s + '})'")
		g.pywrap.Outdent()

		g.pywrap.Printf("def __len__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return _%s_len(self.handle)\n", qNm)
		g.pywrap.Outdent()

		g.pywrap.Printf("def __getitem__(self, key):\n")
		g.pywrap.Indent()
		if ksym.hasHandle() {
			if esym.hasHandle() {
				g.pywrap.Printf("return %s(handle=_%s_elem(self.handle, key.handle))\n", esym.pyPkgId(slc.gopkg), qNm)
			} else {
				g.pywrap.Printf("return _%s_elem(self.handle, key.handle)\n", qNm)
			}
		} else {
			if esym.hasHandle() {
				g.pywrap.Printf("return %s(handle=_%s_elem(self.handle, key))\n", esym.pyPkgId(slc.gopkg), qNm)
			} else {
				g.pywrap.Printf("return _%s_elem(self.handle, key)\n", qNm)
			}
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

		g.pywrap.Printf("def __delitem__(self, key):\n")
		g.pywrap.Indent()
		if ksym.hasHandle() {
			g.pywrap.Printf("return _%s_delete(self.handle, key.handle)\n", qNm)
		} else {
			g.pywrap.Printf("return _%s_delete(self.handle, key)\n", qNm)
		}
		g.pywrap.Outdent()

		g.pywrap.Printf("def keys(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return %s(handle=_%s_keys(self.handle))\n", keyslnm, qNm)
		g.pywrap.Outdent()

		g.pywrap.Printf("def values(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("vls = []\n")
		g.pywrap.Printf("kys = self.keys()\n")
		g.pywrap.Printf("for k in kys:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("vls.append(self[k])\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("return vls\n")
		g.pywrap.Outdent()

		g.pywrap.Printf("def items(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("vls = []\n")
		g.pywrap.Printf("kys = self.keys()\n")
		g.pywrap.Printf("for k in kys:\n")
		g.pywrap.Indent()
		g.pywrap.Printf("vls.append((k, self[k]))\n")
		g.pywrap.Outdent()
		g.pywrap.Printf("return vls\n")
		g.pywrap.Outdent()

		g.pywrap.Printf("def __iter__(self):\n")
		g.pywrap.Indent()
		g.pywrap.Printf("return iter(self.items())\n")
		g.pywrap.Outdent()

		g.pywrap.Printf("def __contains__(self, key):\n")
		g.pywrap.Indent()
		if ksym.hasHandle() {
			g.pywrap.Printf("return _%s_contains(self.handle, key.handle)\n", qNm)
		} else {
			g.pywrap.Printf("return _%s_contains(self.handle, key)\n", qNm)
		}
		g.pywrap.Outdent()

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
		g.gofile.Printf("return CGoHandle(handleFromPtr_%[1]s(&%[2]s{}))\n", slNm, slc.goname)
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s', retval('%s'), [])\n", ctNm, PyHandle)

		// len
		g.gofile.Printf("//export %s_len\n", slNm)
		g.gofile.Printf("func %s_len(handle CGoHandle) int {\n", slNm)
		g.gofile.Indent()
		g.gofile.Printf("return len(deptrFromHandle_%s(handle))\n", slNm)
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_len', retval('int'), [param('%s', 'handle')])\n", slNm, PyHandle)

		// elem
		g.gofile.Printf("//export %s_elem\n", slNm)
		g.gofile.Printf("func %s_elem(handle CGoHandle, _ky %s) %s {\n", slNm, ksym.cgoname, esym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
		if ksym.py2go != "" {
			g.gofile.Printf("v, ok := s[%s(_ky)%s]\n", ksym.py2go, ksym.py2goParenEx)
		} else {
			g.gofile.Printf("v, ok := s[_ky]\n")
		}
		g.gofile.Printf("if !ok {\n")
		g.gofile.Indent()
		g.gofile.Printf("C.PyErr_SetString(C.PyExc_KeyError, C.CString(\"key not in map\"))\n")
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
		if esym.go2py != "" {
			g.gofile.Printf("return %s(v)%s\n", esym.go2py, esym.go2pyParenEx)
		} else {
			g.gofile.Printf("return v\n")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_elem', retval('%s'), [param('%s', 'handle'), param('%s', '_ky')])\n", slNm, esym.cpyname, PyHandle, ksym.cpyname)

		// contains
		g.gofile.Printf("//export %s_contains\n", slNm)
		g.gofile.Printf("func %s_contains(handle CGoHandle, _ky %s) C.char {\n", slNm, ksym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
		if ksym.py2go != "" {
			g.gofile.Printf("_, ok := s[%s(_ky)%s]\n", ksym.py2go, ksym.py2goParenEx)
		} else {
			g.gofile.Printf("_, ok := s[_ky]\n")
		}
		g.gofile.Printf("return boolGoToPy(ok)\n")
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_contains', retval('bool'), [param('%s', 'handle'), param('%s', '_ky')])\n", slNm, PyHandle, ksym.cpyname)

		// set
		g.gofile.Printf("//export %s_set\n", slNm)
		g.gofile.Printf("func %s_set(handle CGoHandle, _ky %s, _vl %s) {\n", slNm, ksym.cgoname, esym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
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

		// delete
		g.gofile.Printf("//export %s_delete\n", slNm)
		g.gofile.Printf("func %s_delete(handle CGoHandle, _ky %s) {\n", slNm, ksym.cgoname)
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
		if ksym.py2go != "" {
			g.gofile.Printf("delete(s, %s(_ky)%s)\n", ksym.py2go, ksym.py2goParenEx)
		} else {
			g.gofile.Printf("delete(s, _ky)\n")
		}
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_delete', None, [param('%s', 'handle'), param('%s', '_ky')])\n", slNm, PyHandle, ksym.cpyname)

		// keys
		g.gofile.Printf("//export %s_keys\n", slNm)
		g.gofile.Printf("func %s_keys(handle CGoHandle) CGoHandle {\n", slNm)
		g.gofile.Indent()
		g.gofile.Printf("s := deptrFromHandle_%s(handle)\n", slNm)
		g.gofile.Printf("kys := make(%s, 0, len(s))\n", keyslsym.goname)
		g.gofile.Printf("for k := range(s) {\n")
		g.gofile.Indent()
		g.gofile.Printf("kys = append(kys, k)\n")
		g.gofile.Outdent()
		g.gofile.Printf("}\n")
		g.gofile.Printf("return %s(&kys)%s\n", keyslsym.go2py, keyslsym.go2pyParenEx)
		g.gofile.Outdent()
		g.gofile.Printf("}\n\n")

		g.pybuild.Printf("mod.add_function('%s_keys', retval('%s'), [param('%s', 'handle')])\n", slNm, keyslsym.cpyname, PyHandle)

	}
}

func (g *pyGen) genMapMethods(s *Map) {
	for _, m := range s.meths {
		g.genMethod(s.sym, m)
	}
}
