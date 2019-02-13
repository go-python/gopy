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

func gopyMakeCmdBuild() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdBuild,
		UsageLine: "build <go-package-name>",
		Short:     "generate and compile (C)Python language bindings for Go",
		Long: `
build generates and compiles (C)Python language bindings for a Go package.

ex:
 $ gopy build [options] <go-package-name>
 $ gopy build github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-build", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for bindings")
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	cmd.Flag.Bool("work", false, "print the name of temporary work directory and do not delete it when exiting")
	return cmd
}

func gopyRunCmdBuild(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) != 1 {
		log.Printf("expect a fully qualified go package name as argument\n")
		return fmt.Errorf(
			"gopy-build: expect a fully qualified go package name as argument",
		)
	}

	var (
		odir    = cmdr.Flag.Lookup("output").Value.Get().(string)
		vm      = cmdr.Flag.Lookup("vm").Value.Get().(string)
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
				"gopy-build: could not create output directory: %v", err,
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
			"gopy-build: go/build.Import failed with path=%q: %v",
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

	err = genPkg(odir, pkg, vm)
	if err != nil {
		return err
	}

	buildname := pkg.Name()
	pkgname := pkg.Name()
	var cmdout []byte
	os.Chdir(odir)

	err = os.Remove(pkgname + ".c")

	fmt.Printf("executing command: go build -buildmode=c-shared ...\n")
	buildname = buildname + "_go"
	cmd = getBuildCommand(odir, buildname, odir, symbols)
	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("executing command: %v build.py\n", vm)
	cmd = exec.Command(vm, "build.py")
	cmdout, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(cmdout))
		return err
	}

	fmt.Printf("executing command: %v-config --cflags\n", vm)
	cmd = exec.Command(vm+"-config", "--cflags") // todo: need minor version!
	cflags, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(cflags))
		return err
	}

	fmt.Printf("executing command: %v-config --ldflags\n", vm)
	cmd = exec.Command(vm+"-config", "--ldflags")
	ldflags, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(ldflags))
		return err
	}

	gccargs := []string{pkgname + ".c", "-dynamiclib", pkgname + "_go" + libExt, "-o", "_" + pkgname + libExt}
	gccargs = append(gccargs, strings.Split(strings.TrimSpace(string(cflags)), " ")...)
	gccargs = append(gccargs, strings.Split(strings.TrimSpace(string(ldflags)), " ")...)

	fmt.Printf("executing command: gcc %v\n", strings.Join(gccargs, " "))
	cmd = exec.Command("gcc", gccargs...)
	cmdout, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v\noutput: %v\n", err, string(cmdout))
		return err
	}

	return err
}

func getBuildCommand(wbuild string, buildname string, work string, symbols bool) (cmd *exec.Cmd) {
	args := []string{"build", "-buildmode=c-shared"}
	if !symbols {
		// These flags will omit the various symbol tables, thereby
		// reducing the final size of the binary. From https://golang.org/cmd/link/
		// -s Omit the symbol table and debug information
		// -w Omit the DWARF symbol table
		args = append(args, "-ldflags=-s -w")
	}
	args = append(args, "-o", filepath.Join(wbuild, buildname)+libExt, ".")
	cmd = exec.Command(
		"go", args...,
	)
	cmd.Dir = work
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
