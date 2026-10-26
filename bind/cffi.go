// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	_ "embed"
)

// The cffi backend (GOPY_BACKEND=cffi) uses the shared plain-C cgo shim (see
// plainc.go), built as a shared library that Python loads with cffi.  Errors
// are recorded for the Python side to raise, instead of being set with
// PyErr_SetString.

//go:embed cffi_build.py
var cffiBuildPy string

// cffiBackend is the cffi backend.  Its callbackABI is in cffi_callback.go.
type cffiBackend struct{}

func (cffiBackend) pyBuildPreamble(g *pyGen) string {
	return g.buildPreamble(cffiBuildPy, "@LIBEXT@", g.libext)
}

func (cffiBackend) genMakefile(g *pyGen, gencmd string, pycfg PyConfig) {
	g.makefile.Printf(MakefileTemplateCFFI, g.cfg.Name, g.cfg.Cmd, gencmd, g.cfg.VM, g.libext, g.cfg.Version)
}
