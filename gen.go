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

func genPkg(odir string, p *bind.Package, lang string) error {
	var err error
	var o *os.File

	switch lang {
	case "python", "py":
		lang, err = getPythonVersion()
		if err != nil {
			return err
		}
	default:
		// no-op
	}

	pyvers := 2
	switch lang {
	case "python2", "py2":
		pyvers = 2
	case "python3", "py3":
		pyvers = 3
	}

	if err != nil {
		return err
	}

	switch lang {
	case "python2", "py2":
		o, err = os.Create(filepath.Join(odir, p.Name()+".c"))
		if err != nil {
			return err
		}
		defer o.Close()
		err = bind.GenCPython(o, fset, p, 2)
		if err != nil {
			return err
		}

	case "python3", "py3":
		return fmt.Errorf("gopy: python-3 support not yet implemented")

	case "go":
		o, err = os.Create(filepath.Join(odir, p.Name()+".go"))
		if err != nil {
			return err
		}
		defer o.Close()

		err = bind.GenGo(o, fset, p, pyvers)
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

func newPackage(path string) (*bind.Package, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	pkg, err := build.Import(path, cwd, 0)
	pkgfiles := make([]string, 0, len(pkg.GoFiles)+len(pkg.CgoFiles))
	pkgfiles = append(pkgfiles, pkg.GoFiles...)
	pkgfiles = append(pkgfiles, pkg.CgoFiles...)
	files, err := parseFiles(pkg.Dir, pkgfiles)
	if err != nil {
		return nil, err
	}

	conf := loader.Config{
		Fset: fset,
	}
	conf.TypeChecker.Error = func(e error) {
		log.Printf("%v\n", e)
		err = e
	}

	p, err := newPackageFrom(files, &conf, pkg)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	return p, err
}

func newPackageFrom(files []*ast.File, conf *loader.Config, pkg *build.Package) (*bind.Package, error) {

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
	if pkgast == nil {
		return nil, fmt.Errorf("gopy: could not find AST for package %q", p.Name())
	}

	pkgdoc := doc.New(pkgast, pkg.ImportPath, 0)

	return bind.NewPackage(p, pkgdoc)
}
