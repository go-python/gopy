// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func gopyMakeCmdBind() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdBind,
		UsageLine: "bind <go-package-name>",
		Short:     "generate and compile (C)Python language bindings for Go",
		Long: `
bind generates and compiles (C)Python language bindings for a Go package.

ex:
 $ gopy bind [options] <go-package-name>
 $ gopy bind github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-bind", flag.ExitOnError),
	}

	cmd.Flag.String("lang", defaultPyVersion, "python version to use for bindings (python2|py2|python3|py3)")
	cmd.Flag.String("output", "", "output directory for bindings")
	return cmd
}

func gopyRunCmdBind(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) != 1 {
		log.Printf("expect a fully qualified go package name as argument\n")
		return fmt.Errorf(
			"gopy-bind: expect a fully qualified go package name as argument",
		)
	}

	odir := cmdr.Flag.Lookup("output").Value.Get().(string)
	lang := cmdr.Flag.Lookup("lang").Value.Get().(string)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if odir == "" {
		odir = cwd
	} else {
		err = os.MkdirAll(odir, 0755)
		if err != nil {
			return fmt.Errorf(
				"gopy-bind: could not create output directory: %v\n", err,
			)
		}
	}
	odir, err = filepath.Abs(odir)
	if err != nil {
		return err
	}

	path := args[0]
	pkg, err := newPackage(path)
	if err != nil {
		return fmt.Errorf(
			"gopy-bind: go/build.Import failed with path=%q: %v\n",
			path,
			err,
		)
	}

	// go-get it to tickle the GOPATH cache (and make sure it compiles
	// correctly)
	cmd := exec.Command(
		"go", "get", "-buildmode=c-shared",
		pkg.ImportPath(),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	work, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create temp-workdir (%v)", err)
	}
	log.Printf("work: %s\n", work)

	err = os.MkdirAll(work, 0644)
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	}
	//defer os.RemoveAll(work)

	err = genPkg(work, pkg, lang)
	if err != nil {
		return err
	}

	err = genPkg(work, pkg, "go")
	if err != nil {
		return err
	}

	wbind, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create temp-workdir (%v)", err)
	}

	err = os.MkdirAll(wbind, 0644)
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	}
	defer os.RemoveAll(wbind)

	buildname := pkg.Name()
	switch lang {
	case "cffi":
		buildname = "_" + buildname
		cmd = exec.Command(
			"go", "build", "-buildmode=c-shared",
			"-o", filepath.Join(wbind, buildname)+".so",
			".",
		)
		cmd.Dir = work
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
		err = processCFFI(pkg.Name(), buildname, odir, work, wbind)
	case "python2", "py2":
		cmd = exec.Command(
			"go", "build", "-buildmode=c-shared",
			"-o", filepath.Join(wbind, buildname)+".so",
			".",
		)
		cmd.Dir = work
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command(
			"/bin/cp",
			filepath.Join(wbind, buildname)+".so",
			filepath.Join(odir, buildname)+".so",
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
	case "python3", "py3":
		return fmt.Errorf("gopy: python-3 support not yet implemented")
	default:
		return fmt.Errorf("unknown target language: %q\n", lang)
	}
	return err
}

func processCFFI(pkgname string, buildname string, odir string, work string, wbind string) error {
	cffipath := filepath.Join(odir, "gopy")
	os.MkdirAll(cffipath, os.ModePerm)

	err := copyFile(filepath.Join(wbind, buildname)+".so", filepath.Join(cffipath, buildname)+".so")
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(work, pkgname)+".py", filepath.Join(odir, pkgname)+".py")
	if err != nil {
		return err
	}

	err = copyFile(filepath.Join(work, "build_"+pkgname)+".py", filepath.Join(odir, "build_"+pkgname)+".py")
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"python",
		filepath.Join(odir, "build_"+pkgname)+".py",
		"",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}
	return err
}
