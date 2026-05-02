// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gilstring is a regression test for the GIL ordering bug (issue #370).
// It mirrors the exact reproduction from the issue report: a string-returning
// function called alongside an integer function from a second extension in the
// same Python process, which triggers crashes under repeated calls.
package gilstring

import "fmt"

// Hello returns a greeting string, mirroring hi.Hello from the issue report.
func Hello(s string) string { return fmt.Sprintf("Hello, %s!", s) }
