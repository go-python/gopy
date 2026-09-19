// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"strings"
	"testing"
)

func TestParseBackend(t *testing.T) {
	for _, tc := range []struct {
		in      string
		want    Backend
		errPart string
	}{
		{in: "", want: BackendPyBindGen},
		{in: "pybindgen", want: BackendPyBindGen},
		{in: " PyBindGen ", want: BackendPyBindGen},
		{in: "cffi", want: BackendCFFI},
		{in: "pybind11", errPart: "not implemented yet"},
		{in: "bogus", errPart: "unknown GOPY_BACKEND"},
	} {
		got, err := parseBackend(tc.in)
		if tc.errPart != "" {
			if err == nil || !strings.Contains(err.Error(), tc.errPart) {
				t.Errorf("parseBackend(%q): got err=%v, want error containing %q", tc.in, err, tc.errPart)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("parseBackend(%q) = %q, %v; want %q", tc.in, got, err, tc.want)
		}
	}
}

func TestBackendFromEnv(t *testing.T) {
	t.Setenv(BackendEnvVar, "")
	if got, err := BackendFromEnv(); err != nil || got != BackendPyBindGen {
		t.Errorf("unset: got %q, %v", got, err)
	}
	t.Setenv(BackendEnvVar, "bogus")
	if _, err := BackendFromEnv(); err == nil {
		t.Error("bogus value: want error")
	}
}
