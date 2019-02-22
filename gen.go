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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goki/gopy/bind"
	"github.com/pkg/errors"
)

var (
	fset = token.NewFileSet()
)

// argStr returns the full command args as a string, without path to exe
func argStr() string {
	ma := make([]string, len(os.Args))
	copy(ma, os.Args)
	_, cmd := filepath.Split(ma[0])
	ma[0] = cmd
	return strings.Join(ma, " ")
}

// genOutDir makes the output directory and returns its absolute path
func genOutDir(odir string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if odir == "" {
		odir = cwd
	} else {
		err = os.MkdirAll(odir, 0755)
		if err != nil {
			return odir, fmt.Errorf("gopy-gen: could not create output directory: %v", err)
		}
	}
	odir, err = filepath.Abs(odir)
	if err != nil {
		return odir, fmt.Errorf("gopy-gen: could not infer absolute path to output directory: %v", err)
	}
	return odir, nil
}

// genPkg generates output for all the current packages that have been parsed,
// in the given output directory
func genPkg(odir, outname, cmdstr, vm string) error {
	var err error
	odir, err = genOutDir(odir)
	if err != nil {
		return err
	}
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
	err = bind.GenPyBind(odir, outname, cmdstr, vm, libExt, pyvers)
	if err != nil {
		log.Println(err)
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
