// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
)

var (
	lang = flag.String("lang", "python", "target language for bindings")
	odir = flag.String("odir", "", "output directory for bindings")
	//pkg  = flag.String("pkg", "", "package name of the bindings")

	usage = `gopy-gen generates Python language bindings for Go.

Usage:

$ gopy-gen [options] <go-package-name>


For usage details, see godoc:

$ godoc github.com/go-python/gopy-gen
`
)

func errorf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, format, args...)
}

func main() {
	var err error

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		log.Printf("expect a fully qualified go package name as argument\n")
		flag.Usage()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if *odir == "" {
		*odir = cwd
	} else {
		err = os.MkdirAll(*odir, 0755)
		if err != nil {
			log.Printf("could not create output directory: %v\n", err)
			os.Exit(1)
		}
	}

	*odir, err = filepath.Abs(*odir)
	if err != nil {
		log.Printf("could not infer absolute path to output directory: %v\n", err)
		os.Exit(1)
	}

	path := flag.Arg(0)
	pkg, err := build.Import(path, cwd, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		os.Exit(1)
	}

	err = genPkg(*odir, pkg)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
