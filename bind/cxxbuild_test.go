// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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
	writeFiles(t, dir, map[string]string{
		"go.mod":      "module m\n\ngo 1.22\n",
		"p/p.go":      "package p\n\n// #cgo LDFLAGS: -L${SRCDIR}/lib -lfoo\nimport \"C\"\n",
		"p/tagged.go": "//go:build gopytag\n\npackage p\n\n// #cgo LDFLAGS: -ltagged\nimport \"C\"\n",
	})
	t.Setenv("CGO_ENABLED", "1")
	for _, tc := range []struct {
		goos, tags string
		pkgs       []string
		want       []string
	}{
		// go list expands ${SRCDIR} with forward slashes, on Windows too
		{runtime.GOOS, "", []string{"./p"}, []string{"-L" + filepath.ToSlash(filepath.Join(dir, "p", "lib")), "-lfoo"}},
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

// writeFiles writes files, named by slash-separated paths relative to dir.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, src := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestMakefileLDFlags checks that the Makefile's link flags come from the
// dependencies its go build will use, which the output directory's go.mod
// picks, rather than those of the module gopy runs in.  Here they differ:
// gopy runs in user, whose go.mod picks depA (-lfromA), while out's go.mod
// picks depB (-lfromB) for the same import path.
func TestMakefileLDFlags(t *testing.T) {
	root := t.TempDir()
	dep := func(lib string) string {
		return "package dep\n\n// #cgo LDFLAGS: -l" + lib + "\nimport \"C\"\n"
	}
	writeFiles(t, root, map[string]string{
		"depA/go.mod": "module example.com/dep\n\ngo 1.22\n",
		"depA/dep.go": dep("fromA"),
		"depB/go.mod": "module example.com/dep\n\ngo 1.22\n",
		"depB/dep.go": dep("fromB"),
		"user/go.mod": "module example.com/user\n\ngo 1.22\n\n" +
			"require example.com/dep v0.0.0\n\n" +
			"replace example.com/dep => ../depA\n",
		"user/p/p.go": "package p\n\nimport _ \"example.com/dep\"\n",
		"out/go.mod": "module out\n\ngo 1.22\n\n" +
			"require (\n\texample.com/dep v0.0.0\n\texample.com/user v0.0.0\n)\n\n" +
			"replace example.com/dep => ../depB\n\n" +
			"replace example.com/user => ../user\n",
	})
	if err := os.MkdirAll(filepath.Join(root, "nomod"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CGO_ENABLED", "1")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")

	// gopy runs in the user's module, as when it's given ./p.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(root, "user")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(cwd) })

	for _, tc := range []struct {
		outDir, want, notWant string
	}{
		{"out", "-lfromB", "-lfromA"},
		// no go.mod in the output directory yet: fall back to gopy's own
		{"nomod", "-lfromA", "-lfromB"},
	} {
		got, err := makefileLDFlags(filepath.Join(root, tc.outDir), "example.com/user/p")
		if err != nil {
			t.Fatalf("makefileLDFlags(%s): %v", tc.outDir, err)
		}
		if !slices.Contains(got, tc.want) || slices.Contains(got, tc.notWant) {
			t.Errorf("makefileLDFlags(%s) = %q, want %s and no %s", tc.outDir, got, tc.want, tc.notWant)
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
// after the archive, where the linker still needs what they resolve, on each
// OS (whose linkers take the archive in different arguments).
func TestCXXArgsLinkOrder(t *testing.T) {
	for _, goos := range []string{"linux", "darwin", "windows"} {
		args := cxxArgs(goos, "p", "_p.so", PyConfig{}, nil, nil, []string{"-lfoo"})
		archive := slices.IndexFunc(args, func(a string) bool { return strings.HasSuffix(a, CXXArchive("p")) })
		lib := slices.Index(args, "-lfoo")
		if archive < 0 || lib < archive {
			t.Errorf("GOOS=%s: CXXArgs = %q, want -lfoo after %s", goos, args, CXXArchive("p"))
		}
	}
}
