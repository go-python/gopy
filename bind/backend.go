// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"os"
	"strings"
)

// BackendEnvVar is the environment variable that selects which tool is used
// to bind the generated cgo shim to CPython.
const BackendEnvVar = "GOPY_BACKEND"

// Backend names a CPython binding tool.
type Backend string

const (
	BackendPyBindGen Backend = "pybindgen" // default
	BackendCFFI      Backend = "cffi"
	BackendPyBind11  Backend = "pybind11"
	BackendNanobind  Backend = "nanobind"
	BackendCAPI      Backend = "capi"
)

// backends lists every known backend.
var backends = []Backend{
	BackendPyBindGen,
	BackendCFFI,
	BackendPyBind11,
	BackendNanobind,
	BackendCAPI,
}

// BackendFromEnv returns the backend selected by GOPY_BACKEND.
// An unset or empty variable selects pybindgen.
func BackendFromEnv() (Backend, error) {
	return parseBackend(os.Getenv(BackendEnvVar))
}

func parseBackend(v string) (Backend, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return BackendPyBindGen, nil
	}
	names := make([]string, len(backends))
	for i, b := range backends {
		names[i] = string(b)
		if string(b) == v {
			return b, nil
		}
	}
	return "", fmt.Errorf("gopy: unknown %s=%q (valid values: %s)", BackendEnvVar, v, strings.Join(names, ", "))
}
