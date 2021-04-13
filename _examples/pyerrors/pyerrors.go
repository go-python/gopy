// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pyerrors holds functions returning an error.
package pyerrors

import (
	"errors"
	"fmt"
)

// Div is a function for detecting errors.
func Div(i, j int) (int, error) {
	if j == 0 {
		return 0, errors.New("Divide by zero.")
	}
	return i / j, nil
}

type Stringer fmt.Stringer
type MyString string

func (t MyString) String() string { return string(t) }

// NewMyString converts a string to a custom MyString type.
// It is an error to pass an empty string value.
func NewMyString(val string) (stringer Stringer, err error) {
	if val == "" {
		err = errors.New("Empty string value.")
		return
	}
	return MyString(val), nil
}
