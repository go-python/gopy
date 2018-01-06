// Copyright 2018 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issue148

import "encoding/json"

type T struct {
	Msg json.RawMessage
}
