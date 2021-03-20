// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// BindCfg is a configuration used during binding generation
type BindCfg struct {
	// output directory for bindings
	OutputDir string
	// name of output package (otherwise name of first package is used)
	Name string
	// code string to run in the go main() function in the cgo library
	Main string
	// the full command args as a string, without path to exe
	Cmd string
	// path to python interpreter
	VM string
	// package prefix used when generating python import statements
	PkgPrefix string
	// rename Go exported symbols to python PEP snake_case
	RenameCase bool
}

// ErrorList is a list of errors
type ErrorList []error

func (list *ErrorList) Add(err error) {
	if err == nil {
		return
	}
	*list = append(*list, err)
}

func (list *ErrorList) Error() error {
	buf := new(bytes.Buffer)
	for i, err := range *list {
		if i > 0 {
			buf.WriteRune('\n')
		}
		io.WriteString(buf, err.Error())
	}
	return errors.New(buf.String())
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
