// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

func (g *pybindGen) genStruct(s Struct) {
	pkgname := s.Package().Name()
	g.pywrap.Printf(`
# Python type for struct %[1]s.%[2]s
class %[2]s:
	""%[3]q""
`,
		pkgname,
		s.GoName(),
		s.Doc(),
	)
	g.pywrap.Indent()
	g.genStructInit(s)
	// g.genStructMembers(s)
	g.genStructMethods(s)
	// g.genStructTPRepr(s)
	// g.genStructTPStr(s)
	g.pywrap.Outdent()
}

func (g *pybindGen) genStructInit(s Struct) {
	// pkg := s.Package()
	// pkgname := s.Package().Name()
	numFields := s.Struct().NumFields()
	numPublic := numFields
	for i := 0; i < s.Struct().NumFields(); i++ {
		f := s.Struct().Field(i)
		if !f.Exported() {
			numPublic--
			continue
		}
	}

	g.pywrap.Printf("def __init__(self, handle):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("self.handle = handle\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")
	g.pywrap.Printf("def __str__(self):\n")
	g.pywrap.Indent()
	g.pywrap.Printf("return self.handle\n")
	g.pywrap.Outdent()
	g.pywrap.Printf("\n")
}

func (g *pybindGen) genStructMethods(s Struct) {
	for _, m := range s.meths {
		g.genMethod(s, m)
	}
}
