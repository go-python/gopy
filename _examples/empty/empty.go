// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package empty does not expose anything.
// We may want to wrap and import it just for its side-effects.
package empty

import (
	"github.com/go-python/gopy/_examples/cpkg"
)

func init() {
	cpkg.Printf("empty.init()... [CALLED]\n")
}
