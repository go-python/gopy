// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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

	cmd.Flag.String("lang", defaultPyVersion, "python version to use for bindings (python2|py2|python3|py3)")
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
	lang := cmdr.Flag.Lookup("lang").Value.Get().(string)

	version := runtime.Version()
	major, minor, err := getGoVersion(version)
	if err != nil {
		return fmt.Errorf("Failed to get the Go version information.")
	}

	if major >= 1 && minor >= 6 {
		godebug := os.Getenv("GODEBUG")
		cgo_check, err := getCgoCheck(godebug)
		if err != nil || cgo_check != 0 {
			return fmt.Errorf("GODEBUG=cgocheck=0 should be set for Go>=1.6")
		}
	}

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
	pkg, err := newPackage(path)
	if err != nil {
		return fmt.Errorf(
			"gopy-bind: go/build.Import failed with path=%q: %v\n",
			path,
			err,
		)
	}

	// go-get it to tickle the GOPATH cache (and make sure it compiles
	// correctly)
	cmd := exec.Command(
		"go", "get", "-buildmode=c-shared",
		pkg.ImportPath(),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
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

	err = genPkg(work, pkg, lang)
	if err != nil {
		return err
	}

	err = genPkg(work, pkg, "go")
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

	buildname := pkg.Name()
	switch lang {
	case "cffi":
		/*
			Since Python importing module priority is XXXX.so > XXXX.py,
			We need to change shared module name from  'XXXX.so' to '_XXXX.so'.
			As the result, an user can import XXXX.py.
		*/
		buildname = "_" + buildname
		cmd = exec.Command(
			"go", "build", "-buildmode=c-shared",
			"-o", filepath.Join(wbind, buildname)+".so",
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
		err = processCFFI(pkg.Name(), buildname, odir, work, wbind)
		if err != nil {
			return err
		}
	case "python2", "py2":
		cmd = exec.Command(
			"go", "build", "-buildmode=c-shared",
			"-o", filepath.Join(wbind, buildname)+".so",
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
			filepath.Join(wbind, buildname)+".so",
			filepath.Join(odir, buildname)+".so",
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
	case "python3", "py3":
		return fmt.Errorf("gopy: python-3 support not yet implemented")
	default:
		return fmt.Errorf("unknown target language: %q\n", lang)
	}
	return err
}

func processCFFI(pkgname string, buildname string, odir string, work string, wbind string) error {
	cmd := exec.Command(
		"/bin/cp",
		filepath.Join(wbind, buildname)+".so",
		filepath.Join(odir, buildname)+".so",
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command(
		"/bin/cp",
		filepath.Join(work, "build_"+pkgname)+".py",
		filepath.Join(odir, "build_"+pkgname)+".py",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command(
		"python",
		filepath.Join(odir, "build_"+pkgname)+".py",
		"",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = odir
	err = cmd.Run()

	if err != nil {
		return err
	}

	// FIXME: This sections should be handled by genPy function.
	wrappedPy, err := ioutil.ReadFile(filepath.Join(work, pkgname) + ".py")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filepath.Join(odir, pkgname)+".py", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(string(wrappedPy))
	if err != nil {
		return err
	}
	return err
}

func getGoVersion(version string) (int64, int64, error) {
	version_regex := regexp.MustCompile(`^go((\d+)(\.(\d+))*)`)
	match := version_regex.FindStringSubmatch(version)
	if match == nil {
		return -1, -1, fmt.Errorf("Invalid Go version information: %s", version)
	}
	version_info := strings.Split(match[1], ".")
	major, _ := strconv.ParseInt(version_info[0], 10, 0)
	minor, _ := strconv.ParseInt(version_info[1], 10, 0)
	return major, minor, nil
}

func getCgoCheck(godebug string) (int, error) {
	if godebug != "" {
		for _, option := range strings.Split(godebug, ",") {
			if strings.HasPrefix(option, "cgocheck=") {
				cgocheck, err := strconv.Atoi(option[9:])
				if err != nil {
					return -1, fmt.Errorf("Invalid cgocheck: %s", option)
				}
				return cgocheck, nil
			}
		}
	}
	return -1, nil
}
