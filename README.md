gopy
====

[![GoDoc](https://godoc.org/github.com/go-python/gopy?status.svg)](https://godoc.org/github.com/go-python/gopy)
[![Build Status](https://travis-ci.org/go-python/gopy.svg?branch=master)](https://travis-ci.org/go-python/gopy)

`gopy` generates (and compiles) a `CPython` extension module from a `go` package.

**WARNING** `gopy` is currently not compatible with `Go>=1.6` and its improved `CGo` rules as documented in [cmd/cgo](https://golang.org/cmd/cgo/#hdr-Passing_pointers).
To be able to run a `CPython` module generated with `Go>=1.6`, one needs to export `GODEBUG=cgocheck=0` to disable the `CGo` rules runtime checker. (see [issue 83](https://github.com/go-python/gopy/issues/83) for more informations.)

## Installation

```sh
$ go get github.com/go-python/gopy
```

You will need `Go >= 1.5`.

## Community

The `go-python` community can be reached out at [go-python@googlegroups.com](mailto:go-python@googlegroups.com) or via the web forum: [go-python group](https://groups.google.com/forum/#!forum/go-python).
See the [CONTRIBUTING](https://github.com/go-python/gopy/blob/master/CONTRIBUTE.md) guide for pointers on how to contribute to `gopy`.

## Documentation

A presentation was given at [dotgo-2015](http://talks.godoc.org/github.com/sbinet/talks/2015/20151109-gopy-dotgo/gopy-dotgo.slide).
A longer version of that talk is also available [here](http://talks.godoc.org/github.com/sbinet/talks/2015/20150929-gopy-lyon/gopy-lyon.slide#17).
An article was also posted on the [GopherAcademy Advent-2015](https://blog.gopheracademy.com/advent-2015/gopy/).

Documentation is available on [godoc](https://godoc.org):
 https://godoc.org/github.com/go-python/gopy
 
or directly from the command-line prompt:

```sh
$ gopy help
gopy - 

Commands:

    bind        generate and compile (C)Python language bindings for Go
    gen         generate (C)Python language bindings for Go

Use "gopy help <command>" for more information about a command.


$ gopy help gen
Usage: gopy gen <go-package-name>

gen generates (C)Python language bindings for a Go package.

ex:
 $ gopy gen [options] <go-package-name>
 $ gopy gen github.com/go-python/gopy/_examples/hi

Options:
  -lang="py2": target language for bindings
  -output="": output directory for bindings


$ gopy help bind
Usage: gopy bind <go-package-name>

bind generates and compiles (C)Python language bindings for a Go package.

ex:
 $ gopy bind [options] <go-package-name>
 $ gopy bind github.com/go-python/gopy/_examples/hi

Options:
  -lang="py2": python version to use for bindings (python2|py2|python3|py3|cffi)
  -output="": output directory for bindings
```


## Examples

### From the `python` shell

`gopy` comes with a little `python` module allowing to wrap and compile `go`
packages directly from the `python` interactive shell:

```python
>>> import gopy
>>> hi = gopy.load("github.com/go-python/gopy/_examples/hi")
gopy> inferring package name...
gopy> loading 'github.com/go-python/gopy/_examples/hi'...
gopy> importing 'github.com/go-python/gopy/_examples/hi'

>>> print hi
<module 'github.com/go-python/gopy/_examples/hi' from '/some/path/.../hi.so'>

>>> print hi.__doc__
package hi exposes a few Go functions to be wrapped and used from Python.
```

### From the command line
```sh
$ gopy bind -output=out github.com/go-python/gopy/_examples/hi
$ ls out
hi.so

$ cd out
$ python2
>>> import hi
>>> dir(hi)
['Add', 'Concat', 'Hello', 'Hi', 'NewPerson', 'Person', '__doc__', '__file__', '__name__', '__package__']

>>> hi.Hello("you")
hello you from go

```

You can also run:

```sh
go test -v -run=TestBind
=== RUN   TestBind
processing "Add"...
processing "Concat"...
processing "Hello"...
processing "Hi"...
processing "NewPerson"...
processing "Person"...
processing "Add"...
processing "Concat"...
processing "Hello"...
processing "Hi"...
processing "NewPerson"...
processing "Person"...
github.com/go-python/gopy/_examples/hi
_/home/binet/dev/go/root/tmp/gopy-431003574
--- hi.Hi()...
hi from go
--- hi.Hello('you')...
hello you from go
--- hi.Add(1, 41)...
42
--- hi.Concat('4', '2')...
42
--- doc(hi.Person):
Person is a simple struct

--- p = hi.Person()...
<hi.Person object at 0x7fc46cc330f0>
['Age', 'Name', '__class__', '__delattr__', '__doc__', '__format__', '__getattribute__', '__hash__', '__init__', '__new__', '__reduce__', '__reduce_ex__', '__repr__', '__setattr__', '__sizeof__', '__str__', '__subclasshook__']
--- p.Name: None
--- p.Age: None
--- doc(p):
Person is a simple struct

--- PASS: TestBind (2.13s)
PASS
ok  	github.com/go-python/gopy	2.135s
```

## Binding generation using Docker (for cross-platform builds)

```
$ cd github.com/go-python/gopy/_examples/hi
$ docker run --rm -v `pwd`:/go/src/in -v `pwd`:/out gopy/gopy app bind -output=/out in
$ file hi.so
hi.so: ELF 64-bit LSB shared object, x86-64, version 1 (SYSV), dynamically linked, not stripped
```

The docker image can also be built on local machine:

```
$ cd $GOPATH/src/github.com/go-python/gopy
$ docker build -t go-python/gopy .
$ docker run -it --rm go-python/gopy
```

## Support Matrix

To know what features are supported on what backends, please refer to the
[Support matrix ](https://github.com/go-python/gopy/blob/master/SUPPORT_MATRIX.md).

## Limitations

- wrap `go` structs into `python` classes **[DONE]**
- better pythonization: turn `go` `errors` into `python` exceptions **[DONE]**
- wrap arrays and slices into types implementing `tp_as_sequence` **[DONE]**
- only `python-2` supported for now **[DONE]**

## Contribute

`gopy` is part of the `go-python` organization and licensed under `BSD-3`.
When you want to contribute a patch or some code to `gopy`, please send a pull
request against the `gopy` issue tracker **AND** a pull request against
[go-python/license](https://github.com/go-python/license) adding yourself to the
`AUTHORS` and `CONTRIBUTORS` files.
