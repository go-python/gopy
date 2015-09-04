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
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-python/gopy/bind"
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

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		bindpkg, err := build.Import("github.com/go-python/gopy/bind", cwd, 0)
		if err != nil {
			return err
		}

		for _, fname := range []string{
			"cgopy_seq_cpy.h",
			"cgopy_seq_cpy.c",
			"cgopy_seq_cpy.go",
		} {
			ftmpl, err := os.Open(filepath.Join(bindpkg.Dir, "_cpy", fname))
			if err != nil {
				return err
			}
			defer ftmpl.Close()

			fout, err := os.Create(filepath.Join(odir, fname))
			if err != nil {
				return err
			}
			defer fout.Close()

			_, err = io.Copy(fout, ftmpl)
			if err != nil {
				return err
			}
			err = fout.Close() // explicit to catch filesystem errors
			if err != nil {
				return err
			}
		}

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

		seqhdr := filepath.Join(odir, "_cgopy_seq_export.h")
		cmd = exec.Command(
			"go", "tool", "cgo",
			"-exportheader", seqhdr,
			filepath.Join(odir, "cgopy_seq_cpy.go"),
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
