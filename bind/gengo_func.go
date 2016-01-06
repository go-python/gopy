// Copyright 2016 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import "fmt"

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
		Descriptor: f.Package().ImportPath() + "." + f.GoName(),
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

	if f.generated {
		g.Printf("cgo_func_%s_(", f.ID())
	} else {
		g.Printf("%s.%s(", g.pkg.Name(), f.GoName())
	}

	for i := range args {
		tail := ""
		if i+1 < len(args) {
			tail = ", "
		}
		g.Printf("_arg_%03d%s", i, tail)
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
	g.Printf("// cgo_func_%[1]s_ wraps read-access to %[2]s.%[3]s\n",
		f.ID(),
		o.Package().Name(),
		o.GoName(),
	)
	g.Printf("func cgo_func_%[1]s_() %[2]s {\n",
		f.ID(),
		sym.GoType().Underlying().String(),
	)
	g.Indent()
	g.Printf("return %s(%s.%s)\n",
		sym.GoType().Underlying().String(),
		o.Package().Name(),
		o.GoName(),
	)
	g.Outdent()
	g.Printf("}\n")
}

func (g *goGen) genFuncSetter(f Func, o Object, sym *symbol) {
	g.Printf("// cgo_func_%[1]s_ wraps write-access to %[2]s.%[3]s\n",
		f.ID(),
		o.Package().Name(),
		o.GoName(),
	)
	g.Printf("func cgo_func_%[1]s_(v %[2]s) {\n",
		f.ID(),
		sym.GoType().Underlying().String(),
	)
	g.Indent()
	g.Printf("%s.%s = %s(v)\n",
		o.Package().Name(),
		o.GoName(),
		sym.gofmt(),
	)
	g.Outdent()
	g.Printf("}\n")
}

func (g *goGen) genFuncNew(f Func, o Object, sym *symbol) {
	g.Printf("// cgo_func_%[1]s_ wraps new-alloc of %[2]s.%[3]s\n",
		f.ID(),
		o.Package().Name(),
		o.GoName(),
	)
	g.Printf("func cgo_func_%[1]s_() *%[2]s {\n",
		f.ID(),
		sym.gofmt(),
	)
	g.Indent()
	g.Printf("var o %[1]s\n", sym.gofmt())
	g.Printf("return &o;\n")
	g.Outdent()
	g.Printf("}\n")
}

func (g *goGen) genFuncTPStr(o Object, sym *symbol, stringer bool) {
	id := o.ID()
	g.Printf("// cgo_func_%[1]s_str wraps Stringer\n", id)
	g.Printf(
		"func cgo_func_%[1]s_str(out, in *seq.Buffer) {\n",
		id,
	)
	g.Indent()
	g.genRead("o", "in", sym.GoType())
	if !stringer {
		g.Printf("str := fmt.Sprintf(\"%%#v\", o)\n")
	} else {
		g.Printf("str := o.String()\n")
	}
	g.Printf("out.WriteString(str)\n")
	g.Outdent()
	g.Printf("}\n")

}
