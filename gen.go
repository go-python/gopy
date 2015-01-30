// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"

	"github.com/go-python/py/bind"
	"golang.org/x/tools/go/loader"
)

var (
	fset = token.NewFileSet()
)

func genPkg(odir string, pkg *build.Package) error {
	var err error
	var o *os.File

	files, err := parseFiles(pkg.Dir, pkg.GoFiles)
	if err != nil {
		return err
	}

	conf := loader.Config{
		SourceImports: true,
		Fset:          fset,
	}
	conf.TypeChecker.Error = func(e error) {
		errorf("%v\n", e)
		err = e
	}

	conf.CreateFromFiles(pkg.ImportPath, files...)
	program, err := conf.Load()
	if err != nil {
		errorf("%v\n", err)
		return err
	}
	p := program.Created[0].Pkg

	switch *lang {
	case "python", "py":
		o, err = os.Create(filepath.Join(odir, p.Name()+".c"))
		if err != nil {
			return err
		}
		err = bind.GenCPython(o, fset, p)
	case "go":
		o, err = os.Create(filepath.Join(odir, p.Name()+".go"))
		if err != nil {
			return err
		}
		err = bind.GenGo(o, fset, p)
	default:
		return fmt.Errorf("unknown target language: %q\n", *lang)
	}

	if err != nil {
		if list, _ := err.(bind.ErrorList); len(list) > 0 {
			for _, err := range list {
				errorf("%v\n", err)
			}
		} else {
			errorf("%v\n", err)
		}
	}

	if err != nil {
		return err
	}

	err = o.Close()
	if err != nil {
		return err
	}

	return err
}

func parseFiles(dir string, fnames []string) ([]*ast.File, error) {
	var (
		files []*ast.File
		err   error
	)

	for _, fname := range fnames {
		path := filepath.Join(dir, fname)
		file, errf := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if errf != nil {
			err = errf
			if list, _ := err.(scanner.ErrorList); len(list) > 0 {
				for _, err := range list {

					errorf("%v\n", err)
				}
			} else {
				errorf("%v\n", err)
			}
		}
		files = append(files, file)
	}

	return files, err
}
