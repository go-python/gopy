// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-python/gopy/bind"
	"golang.org/x/tools/go/loader"
)

var (
	fset = token.NewFileSet()
)

func genPkg(odir string, pkg *build.Package, lang string) error {
	var err error
	var o *os.File

	files, err := parseFiles(pkg.Dir, pkg.GoFiles)
	if err != nil {
		return err
	}

	conf := loader.Config{
		Fset: fset,
	}
	conf.TypeChecker.Error = func(e error) {
		log.Printf("%v\n", e)
		err = e
	}

	p, err := newPackage(files, &conf, pkg)
	if err != nil {
		log.Printf("%v\n", err)
		return err
	}

	switch lang {
	case "python", "py":
		o, err = os.Create(filepath.Join(odir, p.Name()+".c"))
		if err != nil {
			return err
		}
		defer o.Close()
		err = bind.GenCPython(o, fset, p)
		if err != nil {
			return err
		}

	case "go":
		o, err = os.Create(filepath.Join(odir, p.Name()+".go"))
		if err != nil {
			return err
		}
		defer o.Close()

		err = bind.GenGo(o, fset, p)
		if err != nil {
			return err
		}

		var tmpdir string
		tmpdir, err = ioutil.TempDir("", "gopy-go-cgo-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpdir)

		hdr := filepath.Join(odir, p.Name()+".h")
		cmd := exec.Command(
			"go", "tool", "cgo",
			"-exportheader", hdr,
			o.Name(),
		)
		cmd.Dir = tmpdir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown target language: %q\n", lang)
	}

	if err != nil {
		if list, _ := err.(bind.ErrorList); len(list) > 0 {
			for _, err := range list {
				log.Printf("%v\n", err)
			}
		} else {
			log.Printf("%v\n", err)
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

					log.Printf("%v\n", err)
				}
			} else {
				log.Printf("%v\n", err)
			}
		}
		files = append(files, file)
	}

	return files, err
}

func newPackage(files []*ast.File, conf *loader.Config, pkg *build.Package) (*bind.Package, error) {

	conf.CreateFromFiles(pkg.ImportPath, files...)
	program, err := conf.Load()
	if err != nil {
		return nil, err
	}
	p := program.Created[0].Pkg

	var pkgast *ast.Package
	pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	pkgast = pkgs[p.Name()]

	pkgdoc := doc.New(pkgast, pkg.ImportPath, 0)

	return bind.NewPackage(p, pkgdoc)
}
