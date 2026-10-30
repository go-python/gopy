// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cgoldflags needs libraries linked in that its cgo code and the
// standard library ask for: a C library of its own, through #cgo LDFLAGS,
// and (on macOS) the system frameworks crypto/x509 asks for.  The test
// builds clib/libgopyanswer.a from clib/gopyanswer.c first.
package cgoldflags

// #cgo LDFLAGS: -L${SRCDIR}/clib -lgopyanswer
// int gopy_answer(void);
import "C"

import "crypto/x509"

// Answer returns what the C library answers.
func Answer() int {
	return int(C.gopy_answer())
}

// SystemVerify reports whether a certificate that was never parsed verifies
// with the system roots, which it never does.  It links in the system
// verifier, which on macOS calls into the Security framework.
func SystemVerify() bool {
	_, err := new(x509.Certificate).Verify(x509.VerifyOptions{})
	return err == nil
}
