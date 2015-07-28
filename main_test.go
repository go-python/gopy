// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

	want := []byte(`hi from go
hello you from go
--- doc(hi)...
package hi exposes a few Go functions to be wrapped and used from Python.

--- doc(hi.Hi)...
Hi() 

Hi prints hi from Go

--- hi.Hi()...
--- doc(hi.Hello)...
Hello(str s) 

Hello prints a greeting from Go

--- hi.Hello('you')...
--- doc(hi.Add)...
Add(int i, int j) int

Add returns the sum of its arguments.

--- hi.Add(1, 41)...
42
--- hi.Concat('4', '2')...
42
--- doc(hi.Person):
Person is a simple struct

--- p = hi.Person()...
['Age', 'Greet', 'Name', 'String', '__class__', '__delattr__', '__doc__', '__format__', '__getattribute__', '__hash__', '__init__', '__new__', '__reduce__', '__reduce_ex__', '__repr__', '__setattr__', '__sizeof__', '__str__', '__subclasshook__']
--- p.Name: None
--- p.Age: None
--- doc(hi.Greet):
Greet() str

Greet sends greetings

--- p.Greet()...
Hello, I am 
--- doc(p):
Person is a simple struct

`)
	buf := new(bytes.Buffer)
	cmd = exec.Command("python2", "./test.py")
	cmd.Dir = workdir
	cmd.Stdin = os.Stdin
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		t.Fatalf(
			"error running python module: %v\n%v\n", err,
			string(buf.Bytes()),
		)
	}

	if !reflect.DeepEqual(string(buf.Bytes()), string(want)) {
		t.Fatalf("error running python module:\nwant:\n%s\n\ngot:\n%s\n",
			string(want), string(buf.Bytes()),
		)
	}
}
