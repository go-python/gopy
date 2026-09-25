// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	_ "embed"
	"strings"
)

// The pybind11 backend (GOPY_BACKEND=pybind11) shares its cgo shim with cffi
// (see noAPIShim in cffi.go): the same plain-C exported functions, the same
// gopySetError/GopyTakeError error channel, the same byte-slice and callback
// conventions.  What differs is the consumer: instead of a pure-Python module
// that loads the shim at runtime with ctypes, pybind11Build.py writes a C++
// file that #includes the shim's own header and compiles against it directly,
// so (unlike cffi_build.py) it needs no runtime description of the C types --
// the C++ compiler gets them from the header, the same way pybindgen's
// generated .c file does.

//go:embed pybind11_build.py
var pybind11BuildPy string

// pybind11BuildPreamble returns the start of build.py: the pybind11 recorder.
func (g *pyGen) pybind11BuildPreamble() string {
	return strings.NewReplacer(
		"@NAME@", g.cfg.Name,
		"@CMD@", g.cfg.Cmd,
		"@VERSION@", g.cfg.Version,
	).Replace(pybind11BuildPy)
}
