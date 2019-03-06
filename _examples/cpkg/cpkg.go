// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpkg

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//static inline void cpkg_printf(const char *str) {
//	fprintf(stdout, "%s", str);
//	fflush(stdout);
//}
import "C"

import (
	"fmt"
	"unsafe"
)

// Hi prints hi from Go (via C's stdio)
func Hi() {
	cstr := C.CString("hi from go\n")
	defer C.free(unsafe.Pointer(cstr))
	C.cpkg_printf(cstr)
}

// Hello prints a string via C's stdio
func Hello(s string) {
	if s == "" {
		s = "you"
	}
	cstr := C.CString(fmt.Sprintf("hello %s from go\n", s))
	defer C.free(unsafe.Pointer(cstr))
	C.cpkg_printf(cstr)
}

// Printf prints a string via C's stdio
func Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.cpkg_printf(cstr)
}
