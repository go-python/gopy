// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"testing"
)

func TestGetGoVersion(t *testing.T) {
	var versionTests = []struct {
		info  string
		major int64
		minor int64
		err   error
	}{
		{"go1.5", 1, 5, nil},
		{"go1.6", 1, 6, nil},
		{"go1.7", 1, 7, nil},
		{"go1.8", 1, 8, nil},
		{"gcc4", -1, -1, errors.New("Invalid Go version information: gcc4")},
		{"1.8go", -1, -1, errors.New("Invalid Go version information: 1.8go")},
		{"llvm", -1, -1, errors.New("Invalid Go version information: llvm")},
	}

	for _, tt := range versionTests {
		major, minor, err := getGoVersion(tt.info)
		if major != tt.major {
			t.Errorf("getGoVersion(%s): expected major %d, actual %d", tt.info, tt.major, major)
		}

		if minor != tt.minor {
			t.Errorf("getGoVersion(%s): expected major %d, actual %d", tt.info, tt.minor, minor)
		}

		if err != nil && err.Error() != tt.err.Error() {
			t.Errorf("getGoVersion(%s): expected err %s, actual %s", tt.info, tt.err, err)
		}
	}
}

func TestGetCgoCheck(t *testing.T) {
	var cgoTests = []struct {
		info     string
		expected int
		err      error
	}{
		{"", -1, nil},
		{"efence=1", -1, nil},
		{"asdad", -1, nil},
		{"efence=1,cgocheck=1", 1, nil},
		{"efence=1,cgocheck=-1", -1, nil},
		{"cgocheck=2", 2, nil},
		{"cgocheck=asda", -1, errors.New("Invalid cgocheck: cgocheck=asda")},
	}

	for _, tt := range cgoTests {
		actual, err := getCgoCheck(tt.info)
		if actual != tt.expected {
			t.Errorf("getCgoCheck(%s): expected cgocheck %d, actual %d", tt.info, tt.expected, actual)
		}

		if err != nil && err.Error() != tt.err.Error() {
			t.Errorf("getCgoCheck(%s): expected err %s, actual %s", tt.info, tt.err, err)
		}
	}
}
