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

	"github.com/goki/gopy/bind"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
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

ex:
 $ gopy exe [options] <go-package-name> [other-go-package...]
 $ gopy exe github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-exe", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for root of package")
	cmd.Flag.String("name", "", "name of output package (otherwise name of first package is used)")
	cmd.Flag.String("main", "", "code string to run in the go main() function in the cgo library -- defaults to GoPyMainRun() but typically should be overriden")
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

func gopyRunCmdExe(cmdr *commander.Command, args []string) error {
	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	var (
		odir    = cmdr.Flag.Lookup("output").Value.Get().(string)
		name    = cmdr.Flag.Lookup("name").Value.Get().(string)
		mainstr = cmdr.Flag.Lookup("main").Value.Get().(string)
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

	if _, err = os.Stat(setupfn); os.IsNotExist(err) {
		err = GenPyPkgSetup(odir, name, cmdstr, user, version, author, email, desc, url, vm)
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

	odir = filepath.Join(odir, name) // package must be in subdir
	odir, err = genOutDir(odir)
	if err != nil {
		return err
	}

	for _, path := range args {
		var rootdir string
		var err error
		if strings.HasPrefix(path, "./") {
			rootdir = path
		} else {
			rootdir, err = GoSrcDir(path)
		}
		if err != nil {
			return err
		}
		buildPkgRecurse(odir, path, rootdir, rootdir, exmap)
	}
	return runBuild(bind.ModeExe, odir, name, cmdstr, vm, mainstr, symbols)
}
