// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package main

const (
	// libExt = ".dylib"  // theoretically should be this but python only recognizes .so
	libExt       = ".so"
	extraGccArgs = "-dynamiclib"
)
