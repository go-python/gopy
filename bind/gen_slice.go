// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import "go/types"

func (g *pybindGen) genSlice(slc *symbol) {
	pkgname := slc.gopkg.Name()
	// todo: inherit from collections.Iterable too?
	g.pywrap.Printf(`
# Python type for slice %[1]s.%[2]s
class %[2]s(GoClass):
	""%[3]q""
`,
		pkgname,
		slc.id,
		slc.doc,
	)
	g.pywrap.Indent()
	g.genSliceInit(slc)
	// g.genSliceMembers(slc)
	// g.genSliceMethods(slc)
	g.pywrap.Outdent()
}

func (g *pybindGen) genSliceInit(slc *symbol) {
	pkgname := slc.gopkg.Name()
	slNm := slc.id
	qNm := pkgname + "." + slNm
	typ := slc.GoType().Underlying().(*types.Slice)
	esym := g.pkg.syms.symtype(typ.Elem())
	elNm := esym.cgoname

	g.pywrap.Printf("def __init__(self, *args, **kwargs):\n")
	g.pywrap.Indent()
	g.pywrap.Printf(`"""
handle=A Go-side object is always initialized with an explicit handle=arg
otherwise parameter is a python list that we copy from
"""
`)
	g.pywrap.Printf("if len(kwargs) == 1 and 'handle' in kwargs:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = kwargs['handle']\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("else:\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = _%s_CTor()\n", qNm)
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
	g.pywrap.Printf("\n")

	// for _, m := range slc.meths {
	// 	if m.GoName() == "String" {
	// 		g.pywrap.Printf("def __str__(self):\n")
	// 		g.pywrap.Indent()
	// 		g.pywrap.Printf("return self.String()\n")
	// 		g.pywrap.Outdent()
	// 		g.pywrap.Printf("\n")
	// 	}
	// }

	// go ctor
	ctNm := slNm + "_CTor"
	g.gofile.Printf("\n// --- wrapping slice: %v ---\n", qNm)
	g.gofile.Printf("//export %s\n", ctNm)
	g.gofile.Printf("func %s() CGoHandle {\n", ctNm)
	g.gofile.Indent()
	g.gofile.Printf("return handleFmPtr_%[1]s(&%[2]s{})\n", slNm, slc.gofmt())
	g.gofile.Outdent()
	g.gofile.Printf("}\n")

	g.pybuild.Printf("mod.add_function('%s', retval('char*'), [])\n", ctNm)

	// g.pywrap.Printf("	def __repr__(self):
	//        cret = _cffi_helper.lib.cgo_func_0x3243646956_str(self.cgopy)
	//        ret = _cffi_helper.cffi_cgopy_cnv_c2py_string(cret)
	//        return ret

	g.pywrap.Printf("def __len__(self):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("return _%s_len(self.handle)\n", qNm)
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

	g.gofile.Printf("//export %s_len\n", slNm)
	g.gofile.Printf("func %s_len(handle CGoHandle) int {\n", slNm)
	g.gofile.Indent()
	g.gofile.Printf("return len(*ptrFmHandle_%s(handle))\n", slNm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n\n")

	g.pywrap.Printf("def __getitem__(self, idx):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("if idx >= len(self):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("raise IndexError('slice index out of range')\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("return _%s_elem(self.handle, idx)\n", qNm)
	g.pywrap.Outdent()
	g.pywrap.Printf("\n\n")

	g.gofile.Printf("//export %s_elem\n", slNm)
	g.gofile.Printf("func %s_elem(handle CGoHandle, idx int) %s {\n", slNm, elNm)
	g.gofile.Indent()
	g.gofile.Printf("s := *ptrFmHandle_%s(handle)\n", slNm)
	g.gofile.Printf("return %s(s[idx])\n", elNm)
	g.gofile.Outdent()
	g.gofile.Printf("}\n")

	// todo: append?

	g.pywrap.Printf("def __setitem__(self, idx, value):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("if idx >= len(self):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("raise IndexError('slice index out of range')\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("_%s_set(self.handle, idx, value)\n", qNm)
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

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
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")

}
