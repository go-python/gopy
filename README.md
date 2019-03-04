gopy
====

[![GoDoc](https://godoc.org/github.com/go-python/gopy?status.svg)](https://godoc.org/github.com/go-python/gopy)
[![Build Status](https://travis-ci.org/go-python/gopy.svg?branch=master)](https://travis-ci.org/go-python/gopy)
[![Build status](https://ci.appveyor.com/api/projects/status/229rc10wcvsd5t8f?svg=true)](https://ci.appveyor.com/project/sbinet/gopy)

`gopy` generates (and compiles) a `CPython` extension module from a `go` package.

This is a newly-improved version that works with current (e.g., 1.12) versions of Go, and uses unique int64 handles to interface with python, so that no pointers are interchanged, making everything safe for the more recent moving garbage collector.

It also supports python modules having any number of Go packages, and generates a separate .py module file for each package, which link into a single common binding library.  It has been tested extensively on reproducing complex Go code in large libraries -- most stuff "just works".

New features:
* Callback methods from Go into Python now work: you can pass a python function to a Go function that has a function argument, and it will call the python function appropriately.
* The first embedded struct field (i.e., Go's version of type inheritance) is used to establish a corresponding class inheritance in the Python `class` wrappers, which then efficiently inherit all the methods, properties, etc.

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

    gen         generate (C)Python language bindings for Go
    build       generate and compile 
    pkg         generate and compile, automatically including subdirs
                    also creates all the python files needed to install module
    exe         like pkg but makes a standalone executable with go packages bultin
                    this is particularly useful when using -main arg to start process on
                    main thread -- python interpreter can run on another thread.

Use "gopy help <command>" for more information about a command.


$ gopy help gen
Usage: gopy gen <go-package-name> [other-go-package...]

gen generates (C)Python language bindings for Go package(s).

ex:
 $ gopy gen [options] <go-package-name> [other-go-package...]
 $ gopy gen github.com/go-python/gopy/_examples/hi

Options:
  -main="": code string to run in the go main() function in the cgo library
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for bindings
  -vm="python": path to python interpreter

$ gopy help build
Usage: gopy build <go-package-name> [other-go-package...]

build generates and compiles (C)Python language bindings for Go package(s).

ex:
 $ gopy build [options] <go-package-name> [other-go-package...]
 $ gopy build github.com/go-python/gopy/_examples/hi

Options:
  -main="": code string to run in the go main() function in the cgo library
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for bindings
  -symbols=true: include symbols in output
  -vm="python": path to python interpreter

$ gopy help pkg
Usage: gopy pkg <go-package-name> [other-go-package...]

pkg generates and compiles (C)Python language bindings for a Go package, including subdirectories, and generates python module packaging suitable for distribution.  if setup.py file does not yet exist in the target directory, then it along with other default packaging files are created, using arguments.  Typically you create initial default versions of these files and then edit them, and after that, only regenerate the go binding files.

ex:
 $ gopy pkg [options] <go-package-name> [other-go-package...]
 $ gopy pkg github.com/go-python/gopy/_examples/hi

Options:
  -author="gopy": author name
  -desc="": short description of project (long comes from README.md)
  -email="gopy@example.com": author email
  -exclude="": comma-separated list of package names to exclude
  -main="": code string to run in the go GoPyInit() function in the cgo library
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for root of package
  -symbols=true: include symbols in output
  -url="https://github.com/go-python/gopy": home page for project
  -user="": username on https://www.pypa.io/en/latest/ for package name suffix
  -version="0.1.0": semantic version number -- can use e.g., git to get this from tag and pass as argument
  -vm="python": path to python interpreter


$ gopy help exe
Usage: gopy exe <go-package-name> [other-go-package...]

exe generates and compiles (C)Python language bindings for a Go package, including subdirectories, and generates a standalone python executable and associated module packaging suitable for distribution.  if setup.py file does not yet exist in the target directory, then it along with other default packaging files are created, using arguments.  Typically you create initial default versions of these files and then edit them, and after that, only regenerate the go binding files.

The primary need for an exe instead of a pkg dynamic library is when the main thread must be used for something other than running the python interpreter, such as for a GUI library where the main thread must be used for running the GUI event loop (e.g., GoGi).

ex:
 $ gopy exe [options] <go-package-name> [other-go-package...]
 $ gopy exe github.com/go-python/gopy/_examples/hi

Options:
  -author="gopy": author name
  -desc="": short description of project (long comes from README.md)
  -email="gopy@example.com": author email
  -exclude="": comma-separated list of package names to exclude
  -main="": code string to run in the go main() function in the cgo library -- defaults to GoPyMainRun() but typically should be overriden
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for root of package
  -symbols=true: include symbols in output
  -url="https://github.com/go-python/gopy": home page for project
  -user="": username on https://www.pypa.io/en/latest/ for package name suffix
  -version="0.1.0": semantic version number -- can use e.g., git to get this from tag and pass as argument
  -vm="python": path to python interpreter


```


## Examples

### From the `python` shell

NOTE: following not yet working:

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
$ gopy build -output=out github.com/go-python/gopy/_examples/hi
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
- Windows only supported with the `cffi` backend

## Contribute

`gopy` is part of the `go-python` organization and licensed under `BSD-3`.
When you want to contribute a patch or some code to `gopy`, please send a pull
request against the `gopy` issue tracker **AND** a pull request against
[go-python/license](https://github.com/go-python/license) adding yourself to the
`AUTHORS` and `CONTRIBUTORS` files.

