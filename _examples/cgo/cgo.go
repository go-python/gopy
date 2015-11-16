// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cgo tests bindings of CGo-based packages.
package cgo

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//const char* cpkg_sprintf(const char *str) {
//  char *o = (char*)malloc(strlen(str));
//	sprintf(o, "%s", str);
//	return o;
//}
import "C"

import (
	"fmt"
	"unsafe"
)

// Hi returns a string from Go (via C's stdio)
func Hi() string {
	cstr := C.CString("hi from go\n")
	defer C.free(unsafe.Pointer(cstr))
	cout := C.cpkg_sprintf(cstr)
	defer C.free(unsafe.Pointer(cout))
	return C.GoString(cout)
}

// Hello returns a string via C's stdio
func Hello(s string) string {
	if s == "" {
		s = "you"
	}
	cstr := C.CString(fmt.Sprintf("hello %s from go\n", s))
	defer C.free(unsafe.Pointer(cstr))
	cout := C.cpkg_sprintf(cstr)
	defer C.free(unsafe.Pointer(cout))
	return C.GoString(cout)
}
