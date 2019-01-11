// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// getPythonVersion returns the python version available on this machine
func getPythonVersion(vm string) (int, error) {
	py, err := exec.LookPath(vm)
	if err != nil {
		return 0, fmt.Errorf(
			"gopy: could not locate 'python' executable (err: %v)",
			err,
		)
	}

	out, err := exec.Command(py, "-c", "import sys; print(sys.version_info.major)").Output()
	if err != nil {
		return 0, errors.Wrapf(err, "gopy: error retrieving python version")
	}

	vers, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, errors.Wrapf(err, "gopy: error retrieving python version")
	}

	return vers, nil
}
