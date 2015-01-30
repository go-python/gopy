// a go wrapper around py-main
package main

import (
	"os"

	"github.com/go-python/gopy-gen/_examples/py_hi"
	python "github.com/sbinet/go-python" // FIXME(sbinet): migrate to go-python/py
)

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}

	py_hi.Register() // make the python "hi" module available
}

func main() {
	rc := python.Py_Main(os.Args)
	os.Exit(rc)
}
