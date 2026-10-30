// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreamblePlainCStdbool compiles the C part of goPreamblePlainC, which
// cgo copies into <name>_go.h, as C code that #includes it would: before or
// after <stdbool.h>, whose bool must not clash with the preamble's own.
func TestPreamblePlainCStdbool(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("cc not found")
	}
	pre := fmt.Sprintf(goPreamblePlainC, "pkg", "", "", "", "", "", "", "")
	start := strings.Index(pre, "package main\n\n/*\n")
	end := strings.Index(pre, "*/\nimport \"C\"")
	if start < 0 || end < 0 {
		t.Fatal("cgo preamble not found in goPreamblePlainC")
	}
	dir := t.TempDir()
	hdr := filepath.Join(dir, "pkg_go.h")
	if err := os.WriteFile(hdr, []byte(pre[start+len("package main\n\n/*\n"):end]), 0644); err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{
		"#include \"pkg_go.h\"\n#include <stdbool.h>\n",
		"#include <stdbool.h>\n#include \"pkg_go.h\"\n",
	} {
		src += "bool f(bool b) { return !b; }\n"
		file := filepath.Join(dir, "x.c")
		if err := os.WriteFile(file, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		for _, std := range []string{"c99", "c11", "c17"} {
			out, err := exec.Command(cc, "-std="+std, "-fsyntax-only", "-I", dir, file).CombinedOutput()
			if err != nil {
				t.Errorf("-std=%s:\n%s\n%v:\n%s", std, src, err, out)
			}
		}
	}
}
