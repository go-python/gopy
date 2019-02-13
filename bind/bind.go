// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"fmt"
	"go/token"
	"io"
	"os"
)

// ErrorList is a list of errors
type ErrorList []error

func (list ErrorList) Error() string {
	buf := new(bytes.Buffer)
	for i, err := range list {
		if i > 0 {
			buf.WriteRune('\n')
		}
		io.WriteString(buf, err.Error())
	}
	return buf.String()
}

// GenPyBind generates a .go file, build.py file to enable pybindgen to create python bindings,
// and a wrapper .py file that is loaded as the interface to the package with shadow
// python-side classes
func GenPyBind(gow, pybw, pyww, mkw io.Writer, fset *token.FileSet, pkg *Package, vm, libext string, lang int) error {
	gen := &pybindGen{
		gofile:   &printer{buf: new(bytes.Buffer), indentEach: []byte("\t")},
		pybuild:  &printer{buf: new(bytes.Buffer), indentEach: []byte("\t")},
		pywrap:   &printer{buf: new(bytes.Buffer), indentEach: []byte("\t")},
		makefile: &printer{buf: new(bytes.Buffer), indentEach: []byte("\t")},
		fset:     fset,
		pkg:      pkg,
		vm:       vm,
		libext:   libext,
		lang:     lang,
	}
	err := gen.gen()
	if err != nil {
		return err
	}

	_, err = io.Copy(gow, gen.gofile)
	if err != nil {
		return err
	}

	_, err = io.Copy(pybw, gen.pybuild)
	if err != nil {
		return err
	}

	_, err = io.Copy(pyww, gen.pywrap)
	if err != nil {
		return err
	}

	_, err = io.Copy(mkw, gen.makefile)
	if err != nil {
		return err
	}

	return err
}

const (
	doDebug = true
)

func debugf(format string, args ...interface{}) (int, error) {
	if doDebug {
		return fmt.Fprintf(os.Stderr, format, args...)
	}
	return 0, nil
}
