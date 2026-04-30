// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gilstring is a regression test for the GIL ordering bug (issue #370).
// It exercises go2py string conversion through functions, struct fields,
// slice elements, and map values — all of which previously ran C.CString
// without holding the GIL, causing crashes under repeated calls.
package gilstring

import "fmt"

// Add returns the sum of its arguments, mirroring simple.Add from the issue
// report reproduction script.
func Add(i, j int) int { return i + j }

// Hello returns a greeting string, mirroring hi.Hello from the issue report
// reproduction script. Returning a non-trivial string stresses go2py
// string conversion (C.CString) on every call.
func Hello(s string) string { return fmt.Sprintf("Hello, %s!", s) }

// Item is a struct with a string field to exercise struct member string getters.
type Item struct {
	Label string
	Count int
}

// MakeItem returns an Item with the given label.
func MakeItem(s string) Item { return Item{Label: s, Count: len(s)} }

// GetLabel returns the Label field of an Item.
func GetLabel(i Item) string { return i.Label }

// StringSlice returns a slice of strings.
func StringSlice() []string { return []string{"alpha", "beta", "gamma"} }

// StringMap returns a map with string values.
func StringMap() map[string]string { return map[string]string{"key": "value"} }
