// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	_ "embed"
)

// The capi backend (GOPY_BACKEND=capi) is the default backend without
// pybindgen: the same cgo shim, which calls the CPython C API itself, and the
// same build (the generated <name>.c compiled into the Go shared library).
// Only build.py differs: capi_build.py records the pybindgen calls gopy
// writes, and writes <name>.c from them itself.

//go:embed capi_build.py
var capiBuildPy string

func (g *pyGen) isCAPI() bool {
	return g.cfg.Backend == BackendCAPI
}
