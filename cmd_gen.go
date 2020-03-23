// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/go-python/gopy/bind"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func gopyMakeCmdGen() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdGen,
		UsageLine: "gen <go-package-name> [other-go-package...]",
		Short:     "generate (C)Python language bindings for Go",
		Long: `
gen generates (C)Python language bindings for Go package(s).

ex:
 $ gopy gen [options] <go-package-name> [other-go-package...]
 $ gopy gen github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-gen", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for bindings")
	cmd.Flag.String("name", "", "name of output package (otherwise name of first package is used)")
	cmd.Flag.String("main", "", "code string to run in the go main() function in the cgo library")
	cmd.Flag.Bool("no-warn", false, "suppress warning messages, which may be expected")
	return cmd
}

func gopyRunCmdGen(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cmdstr := argStr()

	var (
		odir    = cmdr.Flag.Lookup("output").Value.Get().(string)
		vm      = cmdr.Flag.Lookup("vm").Value.Get().(string)
		name    = cmdr.Flag.Lookup("name").Value.Get().(string)
		mainstr = cmdr.Flag.Lookup("main").Value.Get().(string)
		nowarn  = cmdr.Flag.Lookup("no-warn").Value.Get().(bool)
	)

	if vm == "" {
		vm = "python"
	}

	bind.NoWarn = nowarn

	for _, path := range args {
		pkg, err := newPackage(path)
		if name == "" {
			name = pkg.Name()
		}
		if err != nil {
			return fmt.Errorf("gopy-gen: go/build.Import failed with path=%q: %v", path, err)
		}
	}

	// false = library version
	err = genPkg(bind.ModeGen, odir, name, cmdstr, vm, mainstr)
	if err != nil {
		return err
	}

	return err
}
