// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package empty does not expose anything.
// We may want to wrap and import it just for its side-effects.
package empty

import "fmt"

func init() {
	// todo: not sure why init is not being called!
	fmt.Printf("empty.init()... [CALLED]\n")
}
