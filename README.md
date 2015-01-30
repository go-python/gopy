gopy-gen
========

`gopy-gen` generates a `CPython` extension module from a `go` package.

## Installation

```sh
$ go get github.com/go-python/gopy-gen
```

## Documentation

Documentation is available on [godoc](https://godoc.org):
 https://godoc.org/github.com/go-python/gopy-gen
 
or directly from the command-line prompt:

```sh
$ gopy-gen -help
gopy-gen generates Python language bindings for Go.

Usage:

$ gopy-gen [options] <go-package-name>


For usage details, see godoc:

$ godoc github.com/go-python/gopy-gen
  -lang="python": target language for bindings
  -odir="": output directory for bindings
```


## Examples

```sh
$ gopy-gen -lang=python github.com/go-python/gopy-gen/_examples/hi
$ gopy-gen -lang=go github.com/go-python/gopy-gen/_examples/hi
```

Have also a look at [_examples/py_hi/gen.go](_examples/py_hi/gen.go):
`gopy-gen` can be used via `go generate`.

Running `go generate` in `_examples/py_hi/gen.go` will generate `hi.c`
and `hi.go`.

The `py_hi` package can then be used and imported from a
`go-python`-based `main`.
See [_examples/gopy-test](_examples/gopy-test):

```go
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
```

running `gopy-test`:

```python
>>> import hi
>>> hi.Add(1,2)
3

>>> hi.Hi()
hi from go

>>> hi.Hello("you")
hello you from go

```

## Limitations

- wrap `go` structs into `python` classes
- better pythonization: turn `go` `errors` into `python` exceptions
