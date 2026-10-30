// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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

// TestCgoLDFlags checks that CgoLDFlags finds what a c-archive leaves out:
// a package's own #cgo LDFLAGS, and the standard library's //go:cgo_ldflag
// directives (only darwin has any, which go list finds on any OS).
func TestCgoLDFlags(t *testing.T) {
	dir := t.TempDir()
	for name, src := range map[string]string{
		"go.mod":      "module m\n\ngo 1.22\n",
		"p/p.go":      "package p\n\n// #cgo LDFLAGS: -L${SRCDIR}/lib -lfoo\nimport \"C\"\n",
		"p/tagged.go": "//go:build gopytag\n\npackage p\n\n// #cgo LDFLAGS: -ltagged\nimport \"C\"\n",
	} {
		if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("CGO_ENABLED", "1")
	for _, tc := range []struct {
		goos, tags string
		pkgs       []string
		want       []string
	}{
		{runtime.GOOS, "", []string{"./p"}, []string{"-L" + filepath.Join(dir, "p", "lib"), "-lfoo"}},
		{runtime.GOOS, "gopytag", []string{"./p"}, []string{"-ltagged"}},
		{"darwin", "", []string{"crypto/x509"}, []string{"-framework", "CoreFoundation", "-framework", "Security"}},
		{"darwin", "", []string{"net"}, []string{"-lresolv"}},
		{"linux", "", []string{"crypto/x509"}, nil},
	} {
		t.Setenv("GOOS", tc.goos)
		got, err := CgoLDFlags(dir, tc.tags, tc.pkgs...)
		if err != nil {
			t.Fatalf("GOOS=%s CgoLDFlags(%q, %v): %v", tc.goos, tc.tags, tc.pkgs, err)
		}
		if !containsRun(got, tc.want) {
			t.Errorf("GOOS=%s CgoLDFlags(%q, %v) = %q, want it to have %q", tc.goos, tc.tags, tc.pkgs, got, tc.want)
		}
		if tc.tags == "" && slices.Contains(got, "-ltagged") {
			t.Errorf("GOOS=%s CgoLDFlags(%q, %v) = %q, want no -ltagged", tc.goos, tc.tags, tc.pkgs, got)
		}
		if tc.want == nil && slices.Contains(got, "-framework") {
			t.Errorf("GOOS=%s CgoLDFlags(%q, %v) = %q, want no frameworks", tc.goos, tc.tags, tc.pkgs, got)
		}
	}
}

// containsRun reports whether run appears in s, in order and adjacent.
func containsRun(s, run []string) bool {
	for i := 0; i+len(run) <= len(s); i++ {
		if slices.Equal(s[i:i+len(run)], run) {
			return true
		}
	}
	return false
}

// TestCXXArgsLinkOrder checks that CXXArgs puts the archive's own link flags
// after the archive, where the linker still needs what they resolve.
func TestCXXArgsLinkOrder(t *testing.T) {
	args := CXXArgs("p", "_p.so", PyConfig{}, nil, nil, []string{"-lfoo"})
	archive := slices.Index(args, CXXArchive("p"))
	lib := slices.Index(args, "-lfoo")
	if archive < 0 || lib < archive {
		t.Errorf("CXXArgs = %q, want -lfoo after %s", args, CXXArchive("p"))
	}
}
