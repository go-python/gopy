// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseBackend(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in      string
		want    Backend
		errPart string
	}{
		{in: "", want: BackendPyBindGen},
		{in: "pybindgen", want: BackendPyBindGen},
		{in: " PyBindGen ", want: BackendPyBindGen},
		{in: "cffi", want: BackendCFFI},
		{in: "pybind11", want: BackendPyBind11},
		{in: "nanobind", want: BackendNanobind},
		{in: "capi", want: BackendCAPI},
		{in: "cgo", errPart: "unknown GOPY_BACKEND"},
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

func TestBackendImpl(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		b      Backend
		want   backendImpl
		plainC bool
	}{
		{b: "", want: pybindgenBackend{}},
		{b: BackendPyBindGen, want: pybindgenBackend{}},
		{b: BackendCAPI, want: capiBackend{}},
		{b: BackendCFFI, want: cffiBackend{}, plainC: true},
		{b: BackendPyBind11, want: pybind11Backend, plainC: true},
		{b: BackendNanobind, want: nanobindBackend, plainC: true},
	} {
		got := tc.b.impl()
		if got != tc.want {
			t.Errorf("%q.impl() = %s, want %s", tc.b, implName(got), implName(tc.want))
		}
		// capiBackend embeds pybindgenBackend: it must not pick up a
		// callbackABI that way, nor a C++ backend lose its own.
		if _, plainC := got.(plainCBackend); plainC != tc.plainC {
			t.Errorf("%q.impl() is a plainCBackend: %v, want %v", tc.b, plainC, tc.plainC)
		}
		wantExe := tc.b
		if tc.plainC {
			wantExe = BackendCAPI
		}
		if got := tc.b.ForExe(); got != wantExe {
			t.Errorf("%q.ForExe() = %q, want %q", tc.b, got, wantExe)
		}
		if cb, ok := got.(cxxBackend); ok && cb.name != tc.b {
			t.Errorf("%q.impl().name = %q", tc.b, cb.name)
		}
	}
	for _, b := range backends {
		if b != BackendPyBindGen && b.impl() == (pybindgenBackend{}) {
			t.Errorf("%q.impl() falls back to pybindgen", b)
		}
	}
}

// TestBackendUnknown checks that a Backend none of backends names, which
// BackendFromEnv never returns but a BindCfg can still be given, is refused
// rather than taken for pybindgen.
func TestBackendUnknown(t *testing.T) {
	t.Parallel()
	for _, b := range []Backend{"bogus", "PyBindGen", "cffi "} {
		if got := b.impl(); got != nil {
			t.Errorf("%q.impl() = %s, want nil", b, implName(got))
		}
		// left for GenPyBind to refuse, not swapped for a working backend
		if got := b.ForExe(); got != b {
			t.Errorf("%q.ForExe() = %q, want it unchanged", b, got)
		}
		err := GenPyBind(ModeGen, ".so", 3, false, &BindCfg{Backend: b})
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("unknown backend %q", b)) {
			t.Errorf("GenPyBind with backend %q: err = %v, want unknown backend", b, err)
		}
	}
}

// implName names impl in a test failure, without the embedded build.py a
// cxxBackend's %v would print.
func implName(impl backendImpl) string {
	if cb, ok := impl.(cxxBackend); ok {
		return fmt.Sprintf("cxxBackend(%s)", cb.name)
	}
	return fmt.Sprintf("%T", impl)
}
