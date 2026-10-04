// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	_ "embed"
	"strings"
)

// The nanobind backend (GOPY_BACKEND=nanobind) is the pybind11 backend with
// a different C++ library on the consumer side: the same cgo shim (see
// noAPIShim), the same handle-registry callbacks (see isCXXShim and
// pybind11_callback.go), the same single final link against a c-archive (see
// buildCXXModule in cmd_build.go).  Only nanobind_build.py, a fork of
// pybind11_build.py written against nanobind's API, is its own.

//go:embed nanobind_build.py
var nanobindBuildPy string

// nanobindBuildPreamble returns the start of build.py: the nanobind recorder.
func (g *pyGen) nanobindBuildPreamble() string {
	return strings.NewReplacer(
		"@NAME@", g.cfg.Name,
		"@CMD@", g.cfg.Cmd,
		"@VERSION@", g.cfg.Version,
		"@ARG_HELPERS@", cxxArgHelpers,
	).Replace(nanobindBuildPy)
}
