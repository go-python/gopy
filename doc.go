// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
gopy generates (and compiles) language bindings that make it possible to call Go code
and pass objects from Python.

Using gopy

gopy takes a Go package and generates bindings for all of the exported
symbols. The exported symbols define the cross-language interface.

The gopy tool generates both an API stub in Python, and binding code in
Go. Start with a Go package:

	package hi

	import "fmt"

	func Hello(name string) {
		fmt.Println("Hello, %s!\n", name)
	}

*/
package main
