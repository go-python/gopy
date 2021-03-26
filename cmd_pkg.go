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

	"github.com/go-python/gopy/bind"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

// python packaging links:
// https://pypi.org/
// https://packaging.python.org/tutorials/packaging-projects/
// https://docs.python.org/3/tutorial/modules.html

func gopyMakeCmdPkg() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdPkg,
		UsageLine: "pkg <go-package-name> [other-go-package...]",
		Short:     "generate and compile (C)Python language bindings for Go, and make a python package",
		Long: `
pkg generates and compiles (C)Python language bindings for a Go package, including subdirectories, and generates python module packaging suitable for distribution.  if setup.py file does not yet exist in the target directory, then it along with other default packaging files are created, using arguments.  Typically you create initial default versions of these files and then edit them, and after that, only regenerate the go binding files.

When including multiple packages, list in order of increasing dependency, and use -name arg to give appropriate name.

ex:
 $ gopy pkg [options] <go-package-name> [other-go-package...]
 $ gopy pkg github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-pkg", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for root of package")
	cmd.Flag.String("name", "", "name of output package (otherwise name of first package is used)")
	cmd.Flag.String("main", "", "code string to run in the go GoPyInit() function in the cgo library")
	cmd.Flag.String("package-prefix", ".", "custom package prefix used when generating import "+
		"statements for generated package")
	cmd.Flag.Bool("rename", false, "rename Go symbols to python PEP snake_case")
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	cmd.Flag.String("exclude", "", "comma-separated list of package names to exclude")
	cmd.Flag.String("user", "", "username on https://www.pypa.io/en/latest/ for package name suffix")
	cmd.Flag.String("version", "0.1.0", "semantic version number -- can use e.g., git to get this from tag and pass as argument")
	cmd.Flag.String("author", "gopy", "author name")
	cmd.Flag.String("email", "gopy@example.com", "author email")
	cmd.Flag.String("desc", "", "short description of project (long comes from README.md)")
	cmd.Flag.String("url", "https://github.com/go-python/gopy", "home page for project")
	cmd.Flag.Bool("no-warn", false, "suppress warning messages, which may be expected")
	cmd.Flag.Bool("no-make", false, "do not generate a Makefile, e.g., when called from Makefile")

	return cmd
}

func gopyRunCmdPkg(cmdr *commander.Command, args []string) error {
	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cfg := NewBuildCfg()
	cfg.OutputDir = cmdr.Flag.Lookup("output").Value.Get().(string)
	cfg.Name = cmdr.Flag.Lookup("name").Value.Get().(string)
	cfg.Main = cmdr.Flag.Lookup("main").Value.Get().(string)
	cfg.VM = cmdr.Flag.Lookup("vm").Value.Get().(string)
	cfg.PkgPrefix = cmdr.Flag.Lookup("package-prefix").Value.Get().(string)
	cfg.RenameCase = cmdr.Flag.Lookup("rename").Value.Get().(bool)
	cfg.Symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)
	cfg.NoWarn = cmdr.Flag.Lookup("no-warn").Value.Get().(bool)
	cfg.NoMake = cmdr.Flag.Lookup("no-make").Value.Get().(bool)

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
	return runBuild(bind.ModePkg, cfg)
}

func buildPkgRecurse(odir, path, rootpath string, exmap map[string]struct{}) error {
	buildFirst := path == rootpath
	bpkg, err := loadPackage(path, buildFirst)
	if err != nil {
		return fmt.Errorf("gopy-gen: go build / load of package failed with path=%q: %v", path, err)
	}
	gofiles := bpkg.GoFiles
	onego := ""
	if len(gofiles) == 1 {
		_, onego = filepath.Split(gofiles[0])
	}
	if len(gofiles) == 0 || (len(gofiles) == 1 && onego == "doc.go") {
		fmt.Printf("\n--- skipping dir with no go files or only doc.go: %s -- %s\n", path, gofiles)
		if len(gofiles) == 0 {
			// fmt.Printf("otherfiles: %v\nignorefiles: %v\n", bpkg.OtherFiles, bpkg.IgnoredFiles)
			if len(bpkg.OtherFiles) > 0 {
				gofiles = bpkg.OtherFiles
			} else if len(bpkg.IgnoredFiles) > 0 {
				gofiles = bpkg.IgnoredFiles
			} else {
				return nil // done
			}
		}
	} else {
		// fmt.Printf("gofiles: %s\n", gofiles)
		parsePackage(bpkg)
	}

	//	now try all subdirs
	dir, _ := filepath.Split(gofiles[0])
	drs := Dirs(dir)
	for _, dr := range drs {
		_, ex := exmap[dr]
		if ex || dr[0] == '.' || dr[0] == '_' {
			continue
		}
		sp := filepath.Join(path, dr)
		buildPkgRecurse(odir, sp, rootpath, exmap)
	}
	return nil
}
