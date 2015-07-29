// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestPrinter(t *testing.T) {
	decl := &printer{
		buf:        new(bytes.Buffer),
		indentEach: []byte("\t"),
	}

	impl := &printer{
		buf:        new(bytes.Buffer),
		indentEach: []byte("\t"),
	}

	decl.Printf("\ndecl 1\n")
	decl.Indent()
	decl.Printf(">>> decl-1\n")
	decl.Outdent()
	impl.Printf("impl 1\n")
	decl.Printf("decl 2\n")
	impl.Printf("impl 2\n")

	out := new(bytes.Buffer)
	_, err := io.Copy(out, decl)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	_, err = io.Copy(out, impl)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	want := `
decl 1
	>>> decl-1
decl 2
impl 1
impl 2
`

	str := string(out.Bytes())
	if !reflect.DeepEqual(str, want) {
		t.Fatalf("error:\nwant=%q\ngot =%q\n", want, str)
	}

}
