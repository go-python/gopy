// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pointers

import (
	"github.com/go-python/gopy/_examples/cpkg"
)

//type SPtr *S
type S struct {
	Value int
}

// Inc increments the value of s
func Inc(s *S) {
	cpkg.Printf("==> go: s.Value==%d\n", s.Value)
	s.Value++
	cpkg.Printf("<== go: s.Value==%d\n", s.Value)
}
