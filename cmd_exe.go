// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/rudderlabs/gopy/bind"
)

// python packaging links:
// https://pypi.org/
// https://packaging.python.org/tutorials/packaging-projects/
// https://docs.python.org/3/tutorial/modules.html

func gopyMakeCmdExe() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdExe,
		UsageLine: "exe <go-package-name> [other-go-package...]",
		Short:     "generate and compile (C)Python language bindings for Go, and make a standalone python executable with all the code -- must provide suitable main function code",
		Long: `
exe generates and compiles (C)Python language bindings for a Go package, including subdirectories, and generates a standalone python executable and associated module packaging suitable for distribution.  if setup.py file does not yet exist in the target directory, then it along with other default packaging files are created, using arguments.  Typically you create initial default versions of these files and then edit them, and after that, only regenerate the go binding files.

The primary need for an exe instead of a pkg dynamic library is when the main thread must be used for something other than running the python interpreter, such as for a GUI library where the main thread must be used for running the GUI event loop (e.g., GoGi).

When including multiple packages, list in order of increasing dependency, and use -name arg to give appropriate name.

ex:
 $ gopy exe [options] <go-package-name> [other-go-package...]
 $ gopy exe github.com/rudderlabs/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-exe", flag.ExitOnError),
	}

	AddCommonCmdFlags(&cmd.Flag)

	// Exe specific flags.
	// cmd.Flag.String("package-prefix", ".", "custom package prefix used when generating import "+
	// 	"statements for generated package")
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	cmd.Flag.String("exclude", "", "comma-separated list of package names to exclude")
	cmd.Flag.String("user", "", "username on https://www.pypa.io/en/latest/ for package name suffix")
	cmd.Flag.String("version", "0.1.0", "semantic version number -- can use e.g., git to get this from tag and pass as argument")
	cmd.Flag.String("author", "gopy", "author name")
	cmd.Flag.String("email", "gopy@example.com", "author email")
	cmd.Flag.String("desc", "", "short description of project (long comes from README.md)")
	cmd.Flag.String("url", "https://github.com/rudderlabs/gopy", "home page for project")

	return cmd
}

func gopyRunCmdExe(cmdr *commander.Command, args []string) error {
	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cfg := NewBuildCfg(&cmdr.Flag)
	cfg.Symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)

	var (
		exclude = cmdr.Flag.Lookup("exclude").Value.Get().(string)
		user    = cmdr.Flag.Lookup("user").Value.Get().(string)
		version = cmdr.Flag.Lookup("version").Value.Get().(string)
		author  = cmdr.Flag.Lookup("author").Value.Get().(string)
		email   = cmdr.Flag.Lookup("email").Value.Get().(string)
		desc    = cmdr.Flag.Lookup("desc").Value.Get().(string)
		url     = cmdr.Flag.Lookup("url").Value.Get().(string)
	)

	bind.NoWarn = cfg.NoWarn
	bind.NoMake = cfg.NoMake
	bind.NoPyExceptions = cfg.NoPyExceptions

	if cfg.Name == "" {
		path := args[0]
		_, cfg.Name = filepath.Split(path)
	}

	var err error
	cfg.OutputDir, err = genOutDir(cfg.OutputDir)
	if err != nil {
		return err
	}

	setupfn := filepath.Join(cfg.OutputDir, "setup.py")

	if _, err = os.Stat(setupfn); os.IsNotExist(err) {
		err = GenPyPkgSetup(cfg, user, version, author, email, desc, url)
		if err != nil {
			return err
		}
	}

	defex := []string{"testdata", "internal", "python", "examples", "cmd"}
	excl := append(strings.Split(exclude, ","), defex...)
	exmap := make(map[string]struct{})
	for i := range excl {
		ex := strings.TrimSpace(excl[i])
		exmap[ex] = struct{}{}
	}

	cfg.OutputDir = filepath.Join(cfg.OutputDir, cfg.Name) // package must be in subdir
	cfg.OutputDir, err = genOutDir(cfg.OutputDir)
	if err != nil {
		return err
	}

	for _, path := range args {
		buildPkgRecurse(cfg.OutputDir, path, path, exmap)
	}
	return runBuild(bind.ModeExe, cfg)
}
