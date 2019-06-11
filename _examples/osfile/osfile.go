// Copyright 2019 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package osfile

import (
	"io"
	"log"
	"os"
)

// note: the real test is to access os.File, io.Writer etc directly
// these funcs are just dummies to have something here to compile..

func OpenFile(fname string) *os.File {
	f, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return nil
	}
	return f
}

func WriteToFile(w io.Writer, str string) {
	_, err := w.Write([]byte(str))
	if err != nil {
		log.Println(err)
	}
}
