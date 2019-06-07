// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GoSrcDir tries to locate dir in GOPATH/src/ or GOROOT/src/pkg/ and returns its
// full path. GOPATH may contain a list of paths.  From Robin Elkind github.com/mewkiz/pkg
func GoSrcDir(dir string) (absDir string, err error) {
	for _, srcDir := range build.Default.SrcDirs() {
		absDir = filepath.Join(srcDir, dir)
		finfo, err := os.Stat(absDir)
		if err == nil && finfo.IsDir() {
			return absDir, nil
		}
	}
	return "", fmt.Errorf("kit.GoSrcDir: unable to locate directory (%q) in GOPATH/src/ (%q) or GOROOT/src/pkg/ (%q)", dir, os.Getenv("GOPATH"), os.Getenv("GOROOT"))
}

// Dirs returns a slice of all the directories within a given directory
func Dirs(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}

	var fnms []string
	for _, fi := range files {
		if fi.IsDir() {
			fnms = append(fnms, fi.Name())
		}
	}
	return fnms
}

// GoFiles returns a slice of all the go files within a given directory
func GoFiles(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}

	var fnms []string
	for _, fi := range files {
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go") {
			fnms = append(fnms, fi.Name())
		}
	}
	return fnms
}
