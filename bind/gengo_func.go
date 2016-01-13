// Copyright 2016 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"go/types"
)

func (g *goGen) genFunc(f Func) {
	g.Printf(`
// cgo_func_%[1]s wraps %[2]s.%[3]s
func cgo_func_%[1]s(out, in *seq.Buffer) {
`,
		f.ID(),
		f.Package().Name(),
		f.GoName(),
	)

	g.Indent()
	g.genFuncBody(f)
	g.Outdent()
	g.Printf("}\n\n")

	g.regs = append(g.regs, goReg{
		Descriptor: f.Descriptor(),
		ID:         uhash(f.ID()),
		Func:       f.ID(),
	})
}

func (g *goGen) genFuncBody(f Func) {
	sig := f.Signature()

	args := sig.Params()
	for i, arg := range args {
		g.genRead(fmt.Sprintf("_arg_%03d", i), "in", arg.GoType())
	}

	results := sig.Results()
	if len(results) > 0 {
		for i := range results {
			if i > 0 {
				g.Printf(", ")
			}
			g.Printf("_res_%03d", i)
		}
		g.Printf(" := ")
	}

	if f.typ == nil {
		g.Printf("cgo_func_%s_(", f.ID())
	} else {
		g.Printf("%s.%s(", g.pkg.Name(), f.GoName())
	}

	for i, arg := range args {
		tail := ""
		if i+1 < len(args) {
			tail = ", "
		}
		switch typ := arg.GoType().Underlying().(type) {
		case *types.Struct:
			ptr := types.NewPointer(typ)
			g.Printf("%s%s", g.cnv(typ, ptr, fmt.Sprintf("_arg_%03d", i)), tail)
		default:
			g.Printf("_arg_%03d%s", i, tail)
		}
	}
	g.Printf(")\n")

	if len(results) <= 0 {
		return
	}

	for i, res := range results {
		g.genWrite(fmt.Sprintf("_res_%03d", i), "out", res.GoType())
	}
}

func (g *goGen) genFuncGetter(f Func, o Object, sym *symbol) {
	recv := f.Signature().Recv()
	ret := f.Signature().Results()[0]
	arg := ""
	doc := ""
	get := o.Package().Name() + "." + o.GoName()

	if recv != nil {
		arg = "recv *" + recv.sym.gofmt()
		doc = "." + ret.GoName()
		get = "recv." + ret.GoName()
		if sym.isBasic() {
			get = "*recv"
			doc = ""
		}
	}

	g.Printf("// cgo_func_%[1]s_ wraps read-access to %[2]s.%[3]s%[4]s\n",
		f.ID(),
		o.Package().Name(),
		o.GoName(),
		doc,
	)

	g.Printf("func cgo_func_%[1]s_(%[2]s) %[3]s {\n",
		f.ID(),
		arg,
		ret.sym.gofmt(),
	)
	g.Indent()
	g.Printf("return %s(%s)\n", ret.sym.gofmt(), get)
	g.Outdent()
	g.Printf("}\n\n")
}

func (g *goGen) genFuncSetter(f Func, o Object, sym *symbol) {
	recv := f.Signature().Recv()
	doc := ""
	set := o.Package().Name() + "." + o.GoName()
	typ := sym.gofmt()
	arg := "v " + typ

	if recv != nil {
		fset := f.Signature().Params()[0]
		set = "recv." + fset.GoName()
		doc = "." + fset.GoName()
		typ = fset.sym.gofmt()
		arg = "recv *" + recv.sym.gofmt() + ", v " + typ
		if sym.isBasic() {
			set = "*recv"
			doc = ""
			typ = sym.gofmt()
		}
	}

	g.Printf("// cgo_func_%[1]s_ wraps write-access to %[2]s.%[3]s%[4]s\n",
		f.ID(),
		o.Package().Name(),
		o.GoName(),
		doc,
	)
	g.Printf("func cgo_func_%[1]s_(%[2]s) {\n",
		f.ID(),
		arg,
	)
	g.Indent()
	g.Printf("%s = %s(v)\n", set, typ)
	g.Outdent()
	g.Printf("}\n\n")
}

func (g *goGen) genFuncNew(f Func, typ Type) {
	sym := typ.sym
	g.Printf("// cgo_func_%[1]s_ wraps new-alloc of %[2]s.%[3]s\n",
		f.ID(),
		typ.Package().Name(),
		typ.GoName(),
	)
	if typ := typ.Struct(); typ != nil {
		g.Printf("func cgo_func_%[1]s_() *%[2]s {\n",
			f.ID(),
			sym.gofmt(),
		)
		g.Indent()
		g.Printf("var o %[1]s\n", sym.gofmt())
		g.Printf("return &o;\n")
	} else {
		g.Printf("func cgo_func_%[1]s_() %[2]s {\n",
			f.ID(),
			sym.gofmt(),
		)
		g.Indent()
		g.Printf("var o %[1]s\n", sym.gofmt())
		g.Printf("return o;\n")
	}
	g.Outdent()
	g.Printf("}\n\n")
}

func (g *goGen) genFuncTPStr(typ Type) {
	stringer := typ.prots&ProtoStringer == 1
	sym := typ.sym
	id := typ.ID()
	g.Printf("// cgo_func_%[1]s_str_ wraps Stringer\n", id)
	if typ := typ.Struct(); typ != nil {
		g.Printf(
			"func cgo_func_%[1]s_str_(o *%[2]s) string {\n",
			id,
			sym.gofmt(),
		)
		g.Indent()
		if !stringer {
			g.Printf("str := fmt.Sprintf(\"%%#v\", *o)\n")
		} else {
			g.Printf("str := o.String()\n")
		}
	} else {
		g.Printf(
			"func cgo_func_%[1]s_str_(o %[2]s) string {\n",
			id,
			sym.gofmt(),
		)
		g.Indent()
		if !stringer {
			g.Printf("str := fmt.Sprintf(\"%%#v\", o)\n")
		} else {
			g.Printf("str := o.String()\n")
		}
	}
	g.Printf("return str\n")
	g.Outdent()
	g.Printf("}\n\n")

}
