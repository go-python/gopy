// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
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
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	cmd.Flag.Bool("work", false, "print the name of temporary work directory and do not delete it when exiting")
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
	symbols := cmdr.Flag.Lookup("symbols").Value.Get().(bool)
	printWork := cmdr.Flag.Lookup("work").Value.Get().(bool)

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
				"gopy-bind: could not create output directory: %v", err,
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
			"gopy-bind: go/build.Import failed with path=%q: %v",
			path,
			err,
		)
	}

	// go-get it to tickle the GOPATH cache (and make sure it compiles
	// correctly)
	cmd := exec.Command(
		"go", "get", pkg.ImportPath(),
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
	if printWork {
		log.Printf("work: %s\n", work)
	}

	err = os.MkdirAll(work, 0644)
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	}
	if !printWork {
		defer os.RemoveAll(work)
	}

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
		/*
			Since Python importing module priority is XXXX.so > XXXX.py,
			We need to change shared module name from  'XXXX.so' to '_XXXX.so'.
			As the result, an user can import XXXX.py.
		*/
		buildname = "_" + buildname
		cmd = getBuildCommand(wbind, buildname, work, symbols)
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
		err := cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command(
			"/bin/cp",
			filepath.Join(work, pkg.Name())+".py",
			filepath.Join(odir, pkg.Name())+".py",
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

	case "python2", "py2":
		cmd = getBuildCommand(wbind, buildname, work, symbols)
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
		return fmt.Errorf("gopy: unknown target language: %q", lang)
	}
	return err
}

func getBuildCommand(wbind string, buildname string, work string, symbols bool) (cmd *exec.Cmd) {
	args := []string{"build", "-buildmode=c-shared"}
	if !symbols {
		// These flags will omit the various symbol tables, thereby
		// reducing the final size of the binary. From https://golang.org/cmd/link/
		// -s Omit the symbol table and debug information
		// -w Omit the DWARF symbol table
		args = append(args, "-ldflags=-s -w")
	}
	args = append(args, "-o", filepath.Join(wbind, buildname)+".so", ".")
	cmd = exec.Command(
		"go", args...,
	)
	cmd.Dir = work
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
