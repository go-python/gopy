// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The C library package cgoldflags links in through its #cgo LDFLAGS.  It
// is in a directory of its own, so that cgo doesn't compile it as part of
// the package.
int gopy_answer(void) { return 42; }
