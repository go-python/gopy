// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
)

func (g *cffiGen) genFunc(o Func) {
	sig := o.Signature()
	args := sig.Params()

	funcArgs := []string{}
	for _, arg := range args {
		funcArgs = append(funcArgs, arg.Name())
	}
	g.wrapper.Printf(`
# pythonization of: %[1]s.%[2]s 
def %[2]s(%[3]s):
`,
		g.pkg.pkg.Name(),
		o.GoName(),
		strings.Join(funcArgs, ", "),
	)

	g.wrapper.Indent()
	g.genFuncBody(o)
	g.wrapper.Outdent()
	g.wrapper.Printf("\n")
}

func (g *cffiGen) genFuncBody(f Func) {
	sig := f.Signature()
	res := sig.Results()
	args := sig.Params()
	nres := 0

	funcArgs := []string{}
	for _, arg := range args {
		g.wrapper.Printf("%[1]s = cffi_cnv_py2c_%[2]s(%[3]s)\n", arg.getFuncArg(), arg.GoType(), arg.Name())
		funcArgs = append(funcArgs, arg.getFuncArg())
	}

	if res != nil {
		nres = len(res)
		if nres > 0 {
			g.wrapper.Printf("cret = ")
		}
	}

	g.wrapper.Printf("lib.cgo_func_%[1]s_%[2]s(%[3]s)\n", g.pkg.Name(), f.name, strings.Join(funcArgs, ", "))

	switch nres {
	case 0:
		// no-op
	case 1:
		ret := res[0]
		g.wrapper.Printf("ret = cffi_cnv_c2py_%[1]s(cret)\n", ret.GoType())
		g.wrapper.Printf("return ret\n")
	default:
		panic(fmt.Errorf("gopy: Not yet implemeted for multiple return."))
	}
}
