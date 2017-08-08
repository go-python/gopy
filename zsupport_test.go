// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package main

import (
	"testing"
)

func testPkg(t *testing.T, table pkg) {
	backends := table.lang
	if backends == nil {
		backends = []string{"py2"}
	}
	for _, be := range backends {
		if _, ok := testBackends[be]; !ok {
			// backend not available.
			continue
		}
		switch be {
		case "py2":
			t.Run("py2-python2", func(t *testing.T) {
				t.Parallel()
				testPkgBackend(t, "py2", "python2", table)
			})
		case "py2-cffi":
			t.Run(be, func(t *testing.T) {
				t.Parallel()
				testPkgBackend(t, "cffi", "python2", table)
			})
		case "py3":
			t.Run(be, func(t *testing.T) {
				t.Parallel()
				testPkgBackend(t, be, "python3", table)
			})
		case "py3-cffi":
			t.Run(be, func(t *testing.T) {
				t.Parallel()
				testPkgBackend(t, "cffi", "python3", table)
			})
		case "pypy2-cffi":
			t.Run(be, func(t *testing.T) {
				t.Parallel()
				testPkgBackend(t, "cffi", "pypy", table)
			})
		case "pypy3-cffi":
			t.Run(be, func(t *testing.T) {
				t.Parallel()
				testPkgBackend(t, "cffi", "pypy3", table)
			})
		default:
			t.Errorf("invalid backend name %q", be)
		}
	}
}
