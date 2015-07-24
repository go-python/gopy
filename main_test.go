// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBind(t *testing.T) {
	// mk && rm -rf toto $TMPDIR/gopy-* && gopy bind -output=./toto ./_examples/hi && (echo "=== testing..."; cd toto; cp ../_examples/hi/test.py .; python2 ./test.py && echo "[ok]" || echo "ERR")
	//

	workdir, err := ioutil.TempDir("", "gopy-")
	if err != nil {
		t.Fatalf("could not create workdir: %v\n", err)
	}
	err = os.MkdirAll(workdir, 0644)
	if err != nil {
		t.Fatalf("could not create workdir: %v\n", err)
	}
	defer os.RemoveAll(workdir)

	cmd := exec.Command("gopy", "bind", "-output="+workdir, "./_examples/hi")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatalf("error running gopy-bind: %v\n", err)
	}

	cmd = exec.Command(
		"/bin/cp", "./_examples/hi/test.py",
		filepath.Join(workdir, "test.py"),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatalf("error copying 'test.py': %v\n", err)
	}

	cmd = exec.Command("python2", "./test.py")
	cmd.Dir = workdir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatal("error running python module: %v\n", err)
	}
}
