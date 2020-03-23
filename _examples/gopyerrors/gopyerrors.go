// Copyright 2020 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package gopyerrors contains functions that generate error
// messages from gopy itself.
package gopyerrors

func NotErrorMany() (int, int) {
	return 0, 0
}

func TooMany() (int, int, string) {
	return 0, 1, "Hi"
}

func OK() (int, error) {
	return 0, nil
}

type Struct struct{}

func (s *Struct) NotErrorMany() (int, string) {
	return 0, "Hi"
}

func (s *Struct) TooMany() (int, int, string) {
	return 0, 1, "Hi"
}
