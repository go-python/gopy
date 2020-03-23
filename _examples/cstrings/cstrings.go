// Copyright 2020 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cstrings

import "strings"

func StringValue(s string, n int) string {
	return strings.Repeat(s, n)
}

type StructWithString struct {
	V string
	T int
}

type NestedStructWithString struct {
	A int
	S StructWithString
}

func StringInStruct(s string, n int) StructWithString {
	return StructWithString{
		V: strings.Repeat(s, n),
	}
}

func StringInNestedStruct(s string, n int) NestedStructWithString {
	return NestedStructWithString{
		A: 2,
		S: StructWithString{
			V: strings.Repeat(s, n),
		},
	}
}

func StringSlice(s string, n int) []string {
	return []string{
		strings.Repeat(s, n),
	}
}

func StringMap(s string, n int) map[string]string {
	return map[string]string{
		"a": strings.Repeat(s, n),
	}
}
