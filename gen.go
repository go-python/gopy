// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"

	"github.com/go-python/gopy/bind"
)

// argStr returns the full command args as a string, without path to exe
func argStr() string {
	ma := make([]string, len(os.Args))
	copy(ma, os.Args)
	_, cmd := filepath.Split(ma[0])
	ma[0] = cmd
	for i := range ma {
		if strings.HasPrefix(ma[i], "-main=") {
			ma[i] = "-main=\"" + ma[i][6:] + "\""
		}
	}
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
// mode = gen, build, pkg, exe
func genPkg(mode bind.BuildMode, cfg *BuildCfg) error {
	var err error
	cfg.OutputDir, err = genOutDir(cfg.OutputDir)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(cfg.VM) {
		cfg.VM, err = exec.LookPath(cfg.VM)
		if err != nil {
			return errors.Wrapf(err, "could not locate absolute path to python VM")
		}
	}

	pyvers, err := getPythonVersion(cfg.VM)
	if err != nil {
		return err
	}
	err = bind.GenPyBind(mode, libExt, extraGccArgs, pyvers, cfg.DynamicLinking, &cfg.BindCfg)
	if err != nil {
		log.Println(err)
	}
	return err
}

func loadPackage(path string, buildFirst bool) (*packages.Package, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if buildFirst {
		cmd := exec.Command("go", "build", "-v", path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = cwd

		err = cmd.Run()
		if err != nil {
			log.Printf("Note: there was an error building [%s] -- will continue but it may fail later: %v\n",
				path,
				err,
			)
		}
	}

	// golang.org/x/tools/go/packages supports modules or GOPATH etc
	bpkgs, err := packages.Load(&packages.Config{Mode: packages.LoadTypes}, path)
	if err != nil {
		log.Printf("error resolving import path [%s]: %v\n",
			path,
			err,
		)
		return nil, err
	}

	bpkg := bpkgs[0] // only ever have one at a time
	return bpkg, nil
}

func parsePackage(bpkg *packages.Package) (*bind.Package, error) {
	if len(bpkg.GoFiles) == 0 {
		err := fmt.Errorf("gopy: no files in package %q", bpkg.PkgPath)
		fmt.Println(err)
		return nil, err
	}
	dir, _ := filepath.Split(bpkg.GoFiles[0])
	p := bpkg.Types

	if bpkg.Name == "main" {
		err := fmt.Errorf("gopy: skipping 'main' package %q", bpkg.PkgPath)
		fmt.Println(err)
		return nil, err
	}

	fset := token.NewFileSet()
	var pkgast *ast.Package
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	pkgast = pkgs[p.Name()]
	if pkgast == nil {
		return nil, fmt.Errorf("gopy: could not find AST for package %q", p.Name())
	}

	pkgdoc := doc.New(pkgast, bpkg.PkgPath, 0)
	return bind.NewPackage(p, pkgdoc)
}
