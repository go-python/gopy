// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var makeShellArgTests = []struct {
	arg, want string
}{
	{"", "''"},
	{"-O2", "-O2"},
	{"-I/usr/include/python3.12", "-I/usr/include/python3.12"},
	{"$(CXXLIBFLAGS)", "$(CXXLIBFLAGS)"}, // a make variable, left for make
	{"-Wl,-rpath,$ORIGIN", "'-Wl,-rpath,$$ORIGIN'"},
	{"a$b", "'a$$b'"},
	{"-I/opt/My Python/include", "'-I/opt/My Python/include'"},
	{`C:\hostedtoolcache\Python\include`, `'C:\hostedtoolcache\Python\include'`},
	{"it's", `'it'"'"'s'`},
	{`say "hi"`, `'say "hi"'`},
	{"*.cpp", "'*.cpp'"},
	{"a;b", "'a;b'"},
	{"#x", "'#x'"},
}

func TestMakeShellArg(t *testing.T) {
	for _, tt := range makeShellArgTests {
		if got := makeShellArg(tt.arg); got != tt.want {
			t.Errorf("makeShellArg(%q) = %q, want %q", tt.arg, got, tt.want)
		}
	}
}

// TestMakeShellArgRoundTrip checks makeShellArg's actual contract: written
// into a recipe line, the shell that make runs it with receives arg itself.
func TestMakeShellArgRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not found")
	}
	for _, tt := range makeShellArgTests {
		if tt.arg == "$(CXXLIBFLAGS)" {
			continue // expanded by make, by design
		}
		dir := t.TempDir()
		mk := "all:\n\t@printf '%s' " + makeShellArg(tt.arg) + "\n"
		if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte(mk), 0644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("make", "-s")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("make for %q: %v\n%s", tt.arg, err, out)
			continue
		}
		if string(out) != tt.arg {
			t.Errorf("make passed %q to the shell as %q", tt.arg, out)
		}
	}
}
