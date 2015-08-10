// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpkg

//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//void hello(const char *str) {
//	fprintf(stdout, str);
//}
import "C"

import (
	"fmt"
	"unsafe"
)

// Hello prints a string via C's stdio
func Hello(s string) {
	if s == "" {
		s = "you"
	}
	cstr := C.CString(fmt.Sprintf("hello %s from C\n", s))
	defer C.free(unsafe.Pointer(cstr))
	C.hello(cstr)
}
