// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func gopyMakeCmdGen() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdGen,
		UsageLine: "gen <go-package-name>",
		Short:     "generate (C)Python language bindings for Go",
		Long: `
gen generates (C)Python language bindings for a Go package.

ex:
 $ gopy gen [options] <go-package-name>
 $ gopy gen github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-gen", flag.ExitOnError),
	}

	cmd.Flag.String("lang", defaultPyVersion, "target language for bindings")
	cmd.Flag.String("output", "", "output directory for bindings")
	return cmd
}

func gopyRunCmdGen(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) != 1 {
		log.Printf("expect a fully qualified go package name as argument\n")
		return fmt.Errorf(
			"gopy-gen: expect a fully qualified go package name as argument",
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
				"gopy-gen: could not create output directory: %v", err,
			)
		}
	}

	odir, err = filepath.Abs(odir)
	if err != nil {
		return fmt.Errorf(
			"gopy-gen: could not infer absolute path to output directory: %v",
			err,
		)
	}

	path := args[0]
	pkg, err := newPackage(path)
	if err != nil {
		return fmt.Errorf(
			"gopy-gen: go/build.Import failed with path=%q: %v",
			path,
			err,
		)
	}

	err = genPkg(odir, pkg, lang)
	if err != nil {
		return err
	}

	return err
}
