// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// getPythonVersion returns the python version available on this machine
func getPythonVersion() (string, error) {
	py, err := exec.LookPath("python")
	if err != nil {
		return "", fmt.Errorf(
			"gopy: could not locate 'python' executable (err: %v)",
			err,
		)
	}

	out, err := exec.Command(py, "--version").Output()
	if err != nil {
		return "", fmt.Errorf(
			"gopy: error retrieving python version (err: %v)",
			err,
		)
	}

	vers := ""
	switch {
	case bytes.HasPrefix(out, []byte("Python 2")):
		vers = "py2"
	case bytes.HasPrefix(out, []byte("Python 3")):
		vers = "py3"
	default:
		return "", fmt.Errorf(
			"gopy: invalid python version (%s)",
			string(out),
		)
	}

	return vers, nil
}
