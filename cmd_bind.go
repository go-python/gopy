package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func gopyMakeCmdBind() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdBind,
		UsageLine: "bind <go-package-name>",
		Short:     "generate and compile (C)Python language bindings for Go",
		Long: `
bind generates and compiles (C)Python language bindings for a Go package.

ex:
 $ gopy bind [options] <go-package-name>
 $ gopy bind github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-bind", flag.ExitOnError),
	}

	cmd.Flag.String("output", "", "output directory for bindings")
	return cmd
}

func gopyRunCmdBind(cmdr *commander.Command, args []string) error {
	var err error

	if len(args) != 1 {
		log.Printf("expect a fully qualified go package name as argument\n")
		return fmt.Errorf(
			"gopy-bind: expect a fully qualified go package name as argument",
		)
	}

	odir := cmdr.Flag.Lookup("output").Value.Get().(string)

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
				"gopy-bind: could not create output directory: %v\n", err,
			)
		}
	}
	odir, err = filepath.Abs(odir)
	if err != nil {
		return err
	}

	path := args[0]
	pkg, err := build.Import(path, cwd, 0)
	if err != nil {
		return fmt.Errorf(
			"gopy-bind: go/build.Inport failed with path=%q: %v\n",
			path,
			err,
		)
	}

	work, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create temp-workdir (%v)", err)
	}
	log.Printf("work: %s\n", work)

	err = os.MkdirAll(work, 0644)
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	}
	//defer os.RemoveAll(work)

	cmd := exec.Command(
		"gopy", "gen",
		"-output", work,
		"-lang=python", pkg.ImportPath,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command(
		"gopy", "gen",
		"-output", work,
		"-lang=go", pkg.ImportPath,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	wbind, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create temp-workdir (%v)", err)
	}

	err = os.MkdirAll(wbind, 0644)
	if err != nil {
		return fmt.Errorf("gopy-bind: could not create workdir (%v)", err)
	}
	defer os.RemoveAll(wbind)

	cmd = exec.Command(
		"go", "build", "-v", "-buildmode=c-shared",
		"-o", filepath.Join(wbind, pkg.Name)+".so",
		//	pkg.ImportPath,
		".",
	)
	cmd.Dir = work
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command(
		"/bin/cp",
		filepath.Join(wbind, pkg.Name)+".so",
		filepath.Join(odir, pkg.Name)+".so",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return err
}
