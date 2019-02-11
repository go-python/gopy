// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("api", "pybind", "bindings API to use (pybind, cpython, cffi)")
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

	var (
		odir    = cmdr.Flag.Lookup("output").Value.Get().(string)
		vm      = cmdr.Flag.Lookup("vm").Value.Get().(string)
		api     = cmdr.Flag.Lookup("api").Value.Get().(string)
		symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)
		// printWork = cmdr.Flag.Lookup("work").Value.Get().(bool)
	)

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

	// work, err := ioutil.TempDir("", "gopy-")
	// if err != nil {
	// 	return fmt.Errorf("gopy-bind: could not create temp-workdir (%v)", err)
	// }
	// if printWork {
	// 	log.Printf("work: %s\n", work)
	// }
	//
	// err = os.MkdirAll(work, 0644)
	// if err != nil {
	// 	return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	// }
	// if !printWork {
	// 	defer os.RemoveAll(work)
	// }

	err = genPkg(odir, pkg, vm, api)
	if err != nil {
		return err
	}

	// wbind, err := ioutil.TempDir("", "gopy-")
	// if err != nil {
	// 	return fmt.Errorf("gopy-bind: could not create temp-workdir (%v)", err)
	// }
	//
	// err = os.MkdirAll(wbind, 0644)
	// if err != nil {
	// 	return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	// }
	// defer os.RemoveAll(wbind)

	buildname := pkg.Name()
	pkgname := pkg.Name()
	var cmdout []byte
	switch api {
	case "pybind":

		os.Chdir(odir)

		fmt.Printf("executing command: go build -buildmode=c-shared ...\n")
		buildname = buildname + "_go"
		cmd = getBuildCommand(odir, buildname, odir, symbols)
		err = cmd.Run()
		if err != nil {
			return err
		}

		fmt.Printf("executing command: %v\n", vm+" "+pkgname+".py")
		cmd = exec.Command(vm, pkgname+".py")
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(cmdout))
			return err
		}

		fmt.Printf("executing command: %v.6-config --cflags\n", vm)
		cmd = exec.Command(vm+".6-config", "--cflags") // todo: need minor version!
		cflags, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(cflags))
			return err
		}

		fmt.Printf("executing command: %v.6-config --ldflags\n", vm)
		cmd = exec.Command(vm+".6-config", "--ldflags")
		ldflags, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(ldflags))
			return err
		}

		gccargs := []string{pkgname + ".c", "-dynamiclib", pkgname + "_go" + libExt, "-o", pkgname + libExt}
		gccargs = append(gccargs, strings.Split(strings.TrimSpace(string(cflags)), " ")...)
		gccargs = append(gccargs, strings.Split(strings.TrimSpace(string(ldflags)), " ")...)

		fmt.Printf("executing command: gcc %v\n", strings.Join(gccargs, " "))
		cmd = exec.Command("gcc", gccargs...)
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(cmdout))
			return err
		}

		/*
			case "cffi":
				// Since Python importing module priority is XXXX.so > XXXX.py,
				// We need to change shared module name from  'XXXX.so' to '_XXXX.so'.
				// As the result, an user can import XXXX.py.
				buildname = "_" + buildname
				cmd = getBuildCommand(wbind, buildname, work, symbols)
				err = cmd.Run()
				if err != nil {
					return err
				}

				err = copyCmd(
					filepath.Join(wbind, buildname)+libExt,
					filepath.Join(odir, buildname)+libExt,
				)
				if err != nil {
					return err
				}

				err = copyCmd(
					filepath.Join(work, pkg.Name())+".py",
					filepath.Join(odir, pkg.Name())+".py",
				)
				if err != nil {
					return err
				}

			case "cpython":
				cmd = getBuildCommand(wbind, buildname, work, symbols)
				err = cmd.Run()
				if err != nil {
					return errors.Wrapf(err, "could not generate build command")
				}

				err = copyCmd(
					filepath.Join(wbind, buildname)+libExt,
					filepath.Join(odir, buildname)+libExt,
				)
				if err != nil {
					return err
				}
		*/
	default:
		return fmt.Errorf("gopy: unknown target API: %q", api)
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
	args = append(args, "-o", filepath.Join(wbind, buildname)+libExt, ".")
	cmd = exec.Command(
		"go", args...,
	)
	cmd.Dir = work
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
