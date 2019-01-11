// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/importer"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-python/gopy/bind"
	"github.com/pkg/errors"
)

var (
	fset = token.NewFileSet()
)

func genPkg(odir string, p *bind.Package, vm, capi string) error {
	var err error
	var o *os.File

	if !filepath.IsAbs(vm) {
		vm, err = exec.LookPath(vm)
		if err != nil {
			return errors.Wrapf(err, "could not locate absolute path to python VM")
		}
	}

	pyvers, err := getPythonVersion(vm)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	og, err := os.Create(filepath.Join(odir, p.Name()+".go"))
	if err != nil {
		return err
	}
	defer og.Close()

	err = bind.GenGo(og, fset, p, pyvers, vm, capi)
	if err != nil {
		return err
	}

	switch capi {
	case "cffi":
		o, err = os.Create(filepath.Join(odir, p.Name()+".py"))
		if err != nil {
			return err
		}
		defer o.Close()

		err = bind.GenCFFI(o, fset, p, vm, 2)
		if err != nil {
			return err
		}

	case "cpython":
		o, err = os.Create(filepath.Join(odir, p.Name()+".c"))
		if err != nil {
			return err
		}
		defer o.Close()
		err = bind.GenCPython(o, fset, p, vm, pyvers)
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
			og.Name(),
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
		return fmt.Errorf("gopy: unknown target API: %q", capi)
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

	cmd := exec.Command("go", "install", path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = cwd

	err = cmd.Run()
	if err != nil {
		log.Printf("error installing [%s]: %v\n",
			path,
			err,
		)
		return nil, err
	}

	bpkg, err := build.Import(path, cwd, 0)
	if err != nil {
		log.Printf("error resolving import path [%s]: %v\n",
			path,
			err,
		)
		return nil, err
	}

	pkg, err := importer.Default().Import(bpkg.ImportPath)
	if err != nil {
		log.Printf("error importing package [%v]: %v\n",
			bpkg.ImportPath,
			err,
		)
		return nil, err
	}

	p, err := newPackageFrom(bpkg, pkg)
	if err != nil {
		log.Printf("%v\n", err)
		return nil, err
	}

	return p, err
}

func newPackageFrom(bpkg *build.Package, p *types.Package) (*bind.Package, error) {

	var pkgast *ast.Package
	pkgs, err := parser.ParseDir(fset, bpkg.Dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	pkgast = pkgs[p.Name()]
	if pkgast == nil {
		return nil, fmt.Errorf("gopy: could not find AST for package %q", p.Name())
	}

	pkgdoc := doc.New(pkgast, bpkg.ImportPath, 0)

	return bind.NewPackage(p, pkgdoc)
}
