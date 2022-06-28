// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/rudderlabs/gopy/bind"
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
 $ gopy gen github.com/rudderlabs/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-gen", flag.ExitOnError),
	}

	AddCommonCmdFlags(&cmd.Flag)

	// Gen specific flags.
	cmd.Flag.String("package-prefix", ".", "custom package prefix used when generating import "+
		"statements for generated package")

	return cmd
}

func gopyRunCmdGen(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cfg := NewBuildCfg(&cmdr.Flag)

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
