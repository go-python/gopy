// Copyright 2021 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkgconflict

import (
	"html/template"
	"os"
)

func HtmlTempl(nm string) *template.Template {
	sweaters := Inventory{"hyperlinks", 42}
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
