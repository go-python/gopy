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
	cmd.Flag.String("package-prefix", ".", "custom package prefix used when generating import "+
		"statements for generated package")
	cmd.Flag.Bool("rename", false, "rename Go symbols to python PEP snake_case")
	cmd.Flag.Bool("no-warn", false, "suppress warning messages, which may be expected")
	cmd.Flag.Bool("no-make", false, "do not generate a Makefile, e.g., when called from Makefile")
	return cmd
}

func gopyRunCmdGen(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cfg := NewBuildCfg()
	cfg.OutputDir = cmdr.Flag.Lookup("output").Value.Get().(string)
	cfg.VM = cmdr.Flag.Lookup("vm").Value.Get().(string)
	cfg.Name = cmdr.Flag.Lookup("name").Value.Get().(string)
	cfg.Main = cmdr.Flag.Lookup("main").Value.Get().(string)
	cfg.PkgPrefix = cmdr.Flag.Lookup("package-prefix").Value.Get().(string)
	cfg.RenameCase = cmdr.Flag.Lookup("rename").Value.Get().(bool)
	cfg.NoWarn = cmdr.Flag.Lookup("no-warn").Value.Get().(bool)
	cfg.NoMake = cmdr.Flag.Lookup("no-make").Value.Get().(bool)

	if cfg.VM == "" {
		cfg.VM = "python"
	}

	bind.NoWarn = cfg.NoWarn
	bind.NoMake = cfg.NoMake

	for _, path := range args {
		bpkg, err := loadPackage(path, true) // build first
		if err != nil {
			return fmt.Errorf("gopy-gen: go build / load of package failed with path=%q: %v", path, err)
		}
		pkg, err := parsePackage(bpkg)
		if cfg.Name == "" {
			cfg.Name = pkg.Name()
		}
		if err != nil {
			return err
		}
	}

	err = genPkg(bind.ModeGen, cfg)
	if err != nil {
		return err
	}

	return err
}
