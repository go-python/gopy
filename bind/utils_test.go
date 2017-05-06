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
