// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func init() {

	testEnvironment = os.Environ()

	var (
		py2 = "python2"
		py3 = "python3"
		// pypy2 = "pypy"
		// pypy3 = "pypy3"
	)

	if os.Getenv("GOPY_TRAVIS_CI") == "1" {
		log.Printf("Running in travis CI")
	}

	var (
		disabled []string
		missing  int
	)
	for _, be := range []struct {
		name      string
		vm        string
		module    string
		mandatory bool
	}{
		{"py3", py3, "", true},
		{"py2", py2, "", true},
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
			if be.mandatory {
				missing++
			}
		} else {
			log.Printf("enabling testbackend: %q", be.name)
			testBackends[be.name] = be.vm
		}
	}

	if len(disabled) > 0 {
		log.Printf("The following test backends are not available: %s",
			strings.Join(disabled, ", "))
		if os.Getenv("GOPY_TRAVIS_CI") == "1" && missing > 0 {
			log.Fatalf("Not all backends available in travis CI")
		}
	}
}
