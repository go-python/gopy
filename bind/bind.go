// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"bytes"
	"fmt"
	"go/doc"
	"go/token"
	"io"
	"os"

	"golang.org/x/tools/go/types"
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

// Package ties types.Package and ast.Package together
type Package struct {
	pkg *types.Package
	doc *doc.Package
}

// NewPackage creates a new Package, tying types.Package and ast.Package together.
func NewPackage(pkg *types.Package, doc *doc.Package) *Package {
	return &Package{
		pkg: pkg,
		doc: doc,
	}
}

// Name returns the package name.
func (p *Package) Name() string {
	return p.pkg.Name()
}

// getDoc returns the doc string associated with types.Object
func (p *Package) getDoc(o types.Object) string {
	n := o.Name()
	switch o.(type) {
	case *types.Const:
		for _, c := range p.doc.Consts {
			for _, cn := range c.Names {
				if n == cn {
					return c.Doc
				}
			}
		}

	case *types.Var:
		for _, v := range p.doc.Vars {
			for _, vn := range v.Names {
				if n == vn {
					return v.Doc
				}
			}
		}

	case *types.Func:
		for _, f := range p.doc.Funcs {
			if n == f.Name {
				return f.Doc
			}
		}

	case *types.TypeName:
		for _, t := range p.doc.Types {
			if n == t.Name {
				return t.Doc
			}
		}

	default:
		// TODO(sbinet)
		panic(fmt.Errorf("not yet supported: %v (%T)", o, o))
	}

	return ""
}

// GenCPython generates a (C)Python package from a Go package
func GenCPython(w io.Writer, fset *token.FileSet, pkg *Package) error {
	buf := new(bytes.Buffer)
	gen := &cpyGen{
		decl: &printer{buf: buf, indentEach: []byte("\t")},
		impl: &printer{buf: buf, indentEach: []byte("\t")},
		fset: fset,
		pkg:  pkg,
	}
	err := gen.gen()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, gen.decl.buf)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, gen.impl.buf)
	if err != nil {
		return err
	}

	return err
}

// GenGo generates a cgo package from a Go package
func GenGo(w io.Writer, fset *token.FileSet, pkg *Package) error {
	buf := new(bytes.Buffer)
	gen := &goGen{
		printer: &printer{buf: buf, indentEach: []byte("\t")},
		fset:    fset,
		pkg:     pkg,
	}
	err := gen.gen()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, gen.buf)

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
