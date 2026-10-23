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
// noAPIShim), the same handle-registry callbacks (see cxxBackend and
// pybind11_callback.go), the same single final link against a c-archive (see
// buildCXXModule in cmd_build.go).  Only nanobind_build.py, a fork of
// pybind11_build.py written against nanobind's API, is its own.

//go:embed nanobind_build.py
var nanobindBuildPy string

var nanobindBackend = cxxBackend{
	name:    BackendNanobind,
	buildPy: nanobindBuildPy,
	makeVars: `NANOBIND_INC=$(shell $(PYTHON) -c "import nanobind; print(nanobind.include_dir())")
NANOBIND_SRC=$(shell $(PYTHON) -c "import nanobind; print(nanobind.source_dir())")
# robin_map is a dependency nanobind vendors next to its own headers
CXXLIBFLAGS=-I$(NANOBIND_INC) -I$(NANOBIND_INC)/../ext/robin_map/include ` + strings.Join(NanobindCXXFlags, " ") + `
CXXLIBSRCS=$(NANOBIND_SRC)/nb_combined.cpp
CXXLIBFOUND=$(NANOBIND_INC)
`,
}
