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

	"github.com/goki/ki/dirs"
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

ex:
 $ gopy pkg [options] <go-package-name> [other-go-package...]
 $ gopy pkg github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-pkg", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for root of package")
	cmd.Flag.String("name", "", "name of output package (otherwise name of first package is used)")
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	cmd.Flag.String("exclude", "", "comma-separated list of package names to exclude")
	cmd.Flag.String("user", "", "username on https://www.pypa.io/en/latest/ for package name suffix")
	cmd.Flag.String("version", "0.1.0", "semantic version number -- can use e.g., git to get this from tag and pass as argument")
	cmd.Flag.String("author", "gopy", "author name")
	cmd.Flag.String("email", "gopy@example.com", "author email")
	cmd.Flag.String("desc", "", "short description of project (long comes from README.md)")
	cmd.Flag.String("url", "https://github.com/go-python/gopy", "home page for project")

	return cmd
}

func gopyRunCmdPkg(cmdr *commander.Command, args []string) error {
	if len(args) != 1 {
		log.Printf("expect a fully qualified go package name as argument\n")
		return fmt.Errorf(
			"gopy-pkg: expect a fully qualified go package name as argument",
		)
	}

	var (
		odir    = cmdr.Flag.Lookup("output").Value.Get().(string)
		name    = cmdr.Flag.Lookup("name").Value.Get().(string)
		vm      = cmdr.Flag.Lookup("vm").Value.Get().(string)
		symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)
		exclude = cmdr.Flag.Lookup("exclude").Value.Get().(string)
		user    = cmdr.Flag.Lookup("user").Value.Get().(string)
		version = cmdr.Flag.Lookup("version").Value.Get().(string)
		author  = cmdr.Flag.Lookup("author").Value.Get().(string)
		email   = cmdr.Flag.Lookup("email").Value.Get().(string)
		desc    = cmdr.Flag.Lookup("desc").Value.Get().(string)
		url     = cmdr.Flag.Lookup("url").Value.Get().(string)
	)

	cmdstr := argStr()

	if name == "" {
		path := args[0]
		_, name = filepath.Split(path)
	}

	var err error
	odir, err = genOutDir(odir)
	if err != nil {
		return err
	}

	setupfn := filepath.Join(odir, "setup.py")

	if _, err = os.Stat(string(setupfn)); os.IsNotExist(err) {
		err = GenPyPkgSetup(odir, name, cmdstr, user, version, author, email, desc, url, vm)
		if err != nil {
			return err
		}
	}

	excl := strings.Split(exclude, ",")
	exmap := make(map[string]struct{})
	for i := range excl {
		ex := strings.TrimSpace(excl[i])
		exmap[ex] = struct{}{}
	}

	odir = filepath.Join(odir, name) // package must be in subdir
	odir, err = genOutDir(odir)
	if err != nil {
		return err
	}

	for _, path := range args {
		rootdir, err := dirs.GoSrcDir(path)
		if err != nil {
			return err
		}
		buildPkgRecurse(odir, path, rootdir, rootdir, exmap)
	}
	return runBuild(odir, name, cmdstr, vm, symbols)
}

func buildPkgRecurse(odir, path, rootdir, pathdir string, exmap map[string]struct{}) {
	reldir, _ := filepath.Rel(rootdir, pathdir)
	if reldir == "" {
		newPackage(path)
	} else {
		pkgpath := path + "/" + reldir
		newPackage(pkgpath)
	}

	//	now try all subdirs
	drs := dirs.Dirs(pathdir)
	for _, dr := range drs {
		_, ex := exmap[dr]
		if ex || dr[0] == '.' || dr == "testdata" || dr == "internal" || dr == "python" {
			continue
		}
		sp := filepath.Join(pathdir, dr)
		buildPkgRecurse(odir, path, rootdir, sp, exmap)
	}
}
