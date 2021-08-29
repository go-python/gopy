// Copyright 2021 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkgconflict

import (
	htemplate "html/template"
	"os"
	"text/template"
)

type Inventory struct {
	Material string
	Count    uint
}

func TextTempl(nm string) *template.Template {
	sweaters := Inventory{"wool", 17}
	t, err := template.New(nm).Parse("{{.Count}} items are made of {{.Material}}\n")
	if err != nil {
		panic(err)
	}
	err = t.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}
	return t
}

func HtmlTemplSame(nm string) *htemplate.Template {
	sweaters := Inventory{"fuzz", 2}
	t, err := htemplate.New(nm).Parse("{{.Count}} items are made of {{.Material}}\n")
	if err != nil {
		panic(err)
	}
	err = t.Execute(os.Stdout, sweaters)
	if err != nil {
		panic(err)
	}
	return t
}
