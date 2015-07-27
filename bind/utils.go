// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"golang.org/x/tools/go/types"
)

func isErrorType(typ types.Type) bool {
	return typ == types.Universe.Lookup("error").Type()
}
