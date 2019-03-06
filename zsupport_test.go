// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.7

package main

import (
	"fmt"
	"testing"
)

func testPkg(t *testing.T, table pkg) {
	backends := []string{"py3"} // , "py2"}
	for _, be := range backends {
		fmt.Printf("looping over backends: %s in %s\n", be, backends)
		vm, ok := testBackends[be]
		if !ok || vm == "" {
			// backend not available.
			t.Logf("Skipped testing backend %s for %s\n", be, table.path)
			continue
		}
		switch be {
		// case "py2":
		// 	t.Run("py2-python2", func(t *testing.T) {
		// 		// t.Parallel()
		// 		testPkgBackend(t, vm, table)
		// 	})
		case "py3":
			t.Run(be, func(t *testing.T) {
				// t.Parallel()
				testPkgBackend(t, vm, table)
			})
			return
		default:
			t.Errorf("invalid backend name %q", be)
		}
	}
}
