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

// SupportsExe reports whether gopy exe works with b.  exe mode embeds the
// Python interpreter into the Go binary via the CPython C API (see
// goExePreambleC/Go in gen.go), which a plain-C backend's shim avoids.
func (b Backend) SupportsExe() bool {
	be := b.impl()
	_, plainC := be.(plainCBackend)
	return be != nil && !plainC
}

// backendImpl is what code generation needs from a backend.  The generators
// ask it, rather than testing which Backend is in use.
type backendImpl interface {
	// pyBuildPreamble returns the start of build.py.
	pyBuildPreamble(g *pyGen) string
	// genMakefile writes the Makefile (except for exe mode, which has its
	// own) that runs gencmd and builds the module.
	genMakefile(g *pyGen, gencmd string, pycfg PyConfig)
}

// plainCBackend is a backendImpl whose cgo shim never touches the CPython C
// API (see plainc.go): cffi, pybind11 and nanobind.  A backend that doesn't
// implement it (pybindgen, capi) uses the C API in the shim itself.
type plainCBackend interface {
	backendImpl
	callbackABI
}

var (
	_ backendImpl   = pybindgenBackend{}
	_ backendImpl   = capiBackend{}
	_ plainCBackend = cffiBackend{}
	_ plainCBackend = cxxBackend{}
)

// impl returns the implementation of b, or nil if b is none of backends.  An
// empty Backend is pybindgen, as with an unset GOPY_BACKEND.
func (b Backend) impl() backendImpl {
	switch b {
	case "", BackendPyBindGen:
		return pybindgenBackend{}
	case BackendCAPI:
		return capiBackend{}
	case BackendCFFI:
		return cffiBackend{}
	case BackendPyBind11:
		return pybind11Backend
	case BackendNanobind:
		return nanobindBackend
	}
	return nil
}

// pybindgenBackend is the default backend: build.py drives pybindgen to
// write <name>.c, which cgo compiles into the Go shared library.
type pybindgenBackend struct{}

func (pybindgenBackend) pyBuildPreamble(g *pyGen) string {
	return fmt.Sprintf(PyBuildPreamble, g.cfg.Name, g.cfg.Cmd, g.cfg.Version)
}

func (pybindgenBackend) genMakefile(g *pyGen, gencmd string, pycfg PyConfig) {
	winhack := ""
	if WindowsOS {
		winhack = fmt.Sprintf(`
	# windows-only sed hack here to fix pybindgen declaration of PyInit
	sed -i "s/ PyInit_/ __declspec(dllexport) PyInit_/g" %s.c`, g.cfg.Name)
	}
	g.makefile.Printf(MakefileTemplate, g.cfg.Name, g.cfg.Cmd, gencmd, g.cfg.VM, g.libext,
		ExtModuleName(g.cfg.Name, g.libext, pycfg), winhack, g.cfg.Version)
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
