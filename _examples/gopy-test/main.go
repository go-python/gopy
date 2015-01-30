// a go wrapper around py-main
package main

import (
	"os"

	"github.com/go-python/gopy-gen/_examples/py_hi"
	python "github.com/sbinet/go-python"
)

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}

	py_hi.Import()
}

func main() {
	rc := python.Py_Main(os.Args)
	os.Exit(rc)
}
