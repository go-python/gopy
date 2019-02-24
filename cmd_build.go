// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func gopyMakeCmdBuild() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdBuild,
		UsageLine: "build <go-package-name> [other-go-package...]",
		Short:     "generate and compile (C)Python language bindings for Go",
		Long: `
build generates and compiles (C)Python language bindings for Go package(s).

ex:
 $ gopy build [options] <go-package-name> [other-go-package...]
 $ gopy build github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-build", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for bindings")
	cmd.Flag.String("name", "", "name of output package (otherwise name of first package is used)")
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	return cmd
}

func gopyRunCmdBuild(cmdr *commander.Command, args []string) error {
	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	var (
		odir    = cmdr.Flag.Lookup("output").Value.Get().(string)
		name    = cmdr.Flag.Lookup("name").Value.Get().(string)
		vm      = cmdr.Flag.Lookup("vm").Value.Get().(string)
		symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)
	)

	cmdstr := argStr()

	for _, path := range args {
		pkg, err := newPackage(path)
		if name == "" {
			name = pkg.Name()
		}
		if err != nil {
			return fmt.Errorf("gopy-build: go/build.Import failed with path=%q: %v", path, err)
		}
	}
	return runBuild(odir, name, cmdstr, vm, symbols)
}

func runBuild(odir, outname, cmdstr, vm string, symbols bool) error {
	var err error
	odir, err = genOutDir(odir)
	if err != nil {
		return err
	}
	err = genPkg(odir, outname, cmdstr, vm)
	if err != nil {
		return err
	}

	buildname := outname + "_go"
	var cmdout []byte
	os.Chdir(odir)

	err = os.Remove(outname + ".c")

	fmt.Printf("executing command: goimports -w %v\n", outname+".go")
	cmd := exec.Command("goimports", "-w", outname+".go")
	cmdout, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\no%v\n", err, string(cmdout))
		return err
	}

	args := []string{"build", "-buildmode=c-shared"}
	if !symbols {
		// These flags will omit the various symbol tables, thereby
		// reducing the final size of the binary. From https://golang.org/cmd/link/
		// -s Omit the symbol table and debug information
		// -w Omit the DWARF symbol table
		args = append(args, "-ldflags=-s -w")
	}
	args = append(args, "-o", buildname+libExt, ".")
	fmt.Printf("executing command: go %v\n", strings.Join(args, " "))
	cmd = exec.Command("go", args...)
	cmdout, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
		return err
	}

	fmt.Printf("executing command: %v build.py\n", vm)
	cmd = exec.Command(vm, "build.py")
	cmdout, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\no%v\n", err, string(cmdout))
		return err
	}

	fmt.Printf("executing command: %v-config --cflags\n", vm)
	cmd = exec.Command(vm+"-config", "--cflags") // todo: need minor version!
	cflags, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cflags))
		return err
	}

	fmt.Printf("executing command: %v-config --ldflags\n", vm)
	cmd = exec.Command(vm+"-config", "--ldflags")
	ldflags, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(ldflags))
		return err
	}

	modlib := "_" + outname + libExt
	gccargs := []string{outname + ".c", "-dynamiclib", outname + "_go" + libExt, "-o", modlib}
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
