// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin
// +build darwin

package main

// libExt = ".dylib"  // theoretically should be this but python only recognizes .so
const libExt = ".so"
