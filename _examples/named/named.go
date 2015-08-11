// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package named

type Float float32

func (f Float) Value() float32 { return float32(f) }

type X float32
type XX X
type XXX XX
type XXXX XXX

func (x X) Value() float32    { return float32(x) }
func (x XX) Value() float32   { return float32(x) }
func (x XXX) Value() float32  { return float32(x) }
func (x XXXX) Value() float32 { return float32(x) }
