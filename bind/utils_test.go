// Copyright 2017 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"errors"
	"testing"
)

func TestGetGoVersion(t *testing.T) {
	for _, tt := range []struct {
		info  string
		major int64
		minor int64
		err   error
	}{
		{"go1.5", 1, 5, nil},
		{"go1.6", 1, 6, nil},
		{"go1.7", 1, 7, nil},
		{"go1.8", 1, 8, nil},
		{"gcc4", -1, -1, errors.New("gopy: invalid Go version information: \"gcc4\"")},
		{"1.8go", -1, -1, errors.New("gopy: invalid Go version information: \"1.8go\"")},
		{"llvm", -1, -1, errors.New("gopy: invalid Go version information: \"llvm\"")},
	} {
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

func TestExtractPythonName(t *testing.T) {
	for _, tt := range []struct {
		name     string
		goDoc    string
		newName  string
		newGoDoc string
		err      error
	}{
		{"Func1", "", "Func1", "", nil},
		{"Func2", "\ngopy:name func2\n", "func2", "\n", nil},
		{"Func3", "\ngopy:name bad name\n", "", "", errors.New("gopy: invalid identifier: bad name")},
		{"Func4", "\nsome comment\n", "Func4", "\nsome comment\n", nil},
		{"Func5", "\nsome comment\ngopy:name func5\n", "func5", "\nsome comment\n", nil},
		{"Func6", "\nsome comment\ngopy:name __len__\n", "__len__", "\nsome comment\n", nil},
	} {
		newName, newGoDoc, err := extractPythonName(tt.name, tt.goDoc)

		if newName != tt.newName {
			t.Errorf("extractPythonName(%s, %s): expected name %s, actual %s", tt.name, tt.goDoc, tt.newName, newName)
		}

		if newGoDoc != tt.newGoDoc {
			t.Errorf("extractPythonName(%s, %s): expected comment %s, actual %s", tt.name, tt.goDoc, tt.newGoDoc, newGoDoc)
		}

		if err != nil && err.Error() != tt.err.Error() {
			t.Errorf("extractPythonName(%s, %s): expected err %s, actual %s", tt.name, tt.goDoc, tt.err, err)
		}
	}
}

func TestPythonConfig(t *testing.T) {
	t.Skip()

	for _, tc := range []struct {
		vm   string
		want PyConfig
	}{
		{
			vm: "python2",
		},
		{
			vm: "python3",
		},
	} {
		t.Run(tc.vm, func(t *testing.T) {
			cfg, err := GetPythonConfig(tc.vm)
			if err != nil {
				t.Fatal(err)
			}
			if cfg != tc.want {
				t.Fatalf("error:\ngot= %#v\nwant=%#v\n", cfg, tc.want)
			}
		})
	}
}
