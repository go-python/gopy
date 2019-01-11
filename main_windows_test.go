// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func init() {
	if os.Getenv("GOPY_APPVEYOR_CI") == "1" {
		log.Printf("Running in appveyor CI")
	}

	var (
		cpy2dir = os.Getenv("CPYTHON2DIR")
		cpy3dir = os.Getenv("CPYTHON3DIR")
	)

	var disabled []string
	for _, be := range []struct {
		name   string
		vm     string
		module string
	}{
		{"py2", path.Join(cpy2dir, "python"), ""},
		{"py2-cffi", path.Join(cpy2dir, "python"), "cffi"},
		{"py3", path.Join(cpy3dir, "python"), ""},
		{"py3-cffi", path.Join(cpy3dir, "python"), "cffi"},
		//	{"pypy2-cffi", "pypy", "cffi"},
		//	{"pypy3-cffi", "pypy3", "cffi"},
	} {
		args := []string{"-c", ""}
		if be.module != "" {
			args[1] = "import " + be.module
		}
		log.Printf("checking testbackend: %q...", be.name)
		cmd := exec.Command(be.vm, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("disabling testbackend: %q, error: '%s'", be.name, err.Error())
			testBackends[be.name] = ""
			disabled = append(disabled, be.name)
		} else {
			log.Printf("enabling testbackend: %q", be.name)
			testBackends[be.name] = be.vm
		}
	}

	if len(disabled) > 0 {
		log.Printf("The following test backends are not available: %s",
			strings.Join(disabled, ", "))
		if os.Getenv("GOPY_APPVEYOR_CI") == "1" {
			log.Fatalf("Not all backends available in appveyor CI")
		}
	}
}
