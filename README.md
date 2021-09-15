gopy
====

[![GoDoc](https://godoc.org/github.com/go-python/gopy?status.svg)](https://godoc.org/github.com/go-python/gopy)
[![CI](https://github.com/go-python/gopy/workflows/CI/badge.svg)](https://github.com/go-python/gopy/actions)
[![Build status](https://ci.appveyor.com/api/projects/status/229rc10wcvsd5t8f?svg=true)](https://ci.appveyor.com/project/sbinet/gopy)

`gopy` generates (and compiles) a `CPython` extension module from a `go` package.

This is an improved version that works with current versions of Go (e.g., 1.15 -- should work with any future version going forward), and uses unique int64 handles to interface with python, so that no pointers are interchanged, making everything safe for the more recent moving garbage collector.

It also supports python modules having any number of Go packages, and generates a separate .py module file for each package, which link into a single common binding library.  It has been tested extensively on reproducing complex Go code in large libraries -- most stuff "just works".  For example, the [GoGi](https://github.com/goki/gi) GUI library is fully usable from python now (do `make; make install` in the python directory there, and try the `examples/widgets/widgets.py` demo).

New features:
* Callback methods from Go into Python now work: you can pass a python function to a Go function that has a function argument, and it will call the python function appropriately.
* The first embedded struct field (i.e., Go's version of type inheritance) is used to establish a corresponding class inheritance in the Python `class` wrappers, which then efficiently inherit all the methods, properties, etc.

## Installation

Gopy now assumes that you are working with modules-based builds, and requires a valid `go.mod` file, and works only with Go versions 1.15 and above.

Currently using [pybindgen](https://pybindgen.readthedocs.io/en/latest/tutorial/) to generate the low-level c-to-python bindings, but support for [cffi](https://cffi.readthedocs.io/en/latest/) should be relatively straightforward for those using PyPy instead of CPython (pybindgen should be significantly faster for CPython apparently).  You also need `goimports` to ensure the correct imports are included.

```sh
$ python3 -m pip install pybindgen
$ go get golang.org/x/tools/cmd/goimports
$ go get github.com/go-python/gopy
```

(This all assumes you have already installed [Go itself](https://golang.org/doc/install), and added `~/go/bin` to your `PATH`).

To [install python modules](https://packaging.python.org/tutorials/packaging-projects/), you will need the python install packages:

```sh
python3 -m pip install --upgrade setuptools wheel
```

IMPORTANT: many errors will be avoided by specifying the `-vm` option to gopy, with a full path if needed, or typically just `-vm=python3` to use python3 instead of version 2, which is often the default for the plain `python` command.

### Linux

On linux, you may need to ensure that the linker `ld` will look in the current directory for library files -- add this to your `.bashrc` file (and `source` that file after editing, or enter command locally):

```sh
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.
```

### Windows

As of version 0.4.0, windows is now better supported, and is passing tests (on at least one developers machine).  You may still need to set some environment variables depending on your python installation, but a vanilla standard install is working.

Install Python from the main Python distribution: https://www.python.org/downloads/windows/ -- *do not install from the Microsoft Store app!* -- while that is very convenient, it creates symbolic links to access the python executables, which is incompatible with go exec.Command to run it, despite too many hours of trying to get around that.

The standard python install does not create a `python3.exe` which gopy looks for -- follow instructions here:
https://stackoverflow.com/questions/39910730/python3-is-not-recognized-as-an-internal-or-external-command-operable-program/41492852
(just make a copy of python.exe to python3.exe in the relevant installed location).

If you get a bunch of errors during linking in the build process, set `LIBDIR` or `GOPY_LIBDIR` to path to python libraries, and `LIBRARY` or `GOPY_PYLIB` to name of python library (e.g., python39 for 3.9).

## Community

See the [CONTRIBUTING](https://github.com/go-python/gopy/blob/master/CONTRIBUTE.md) guide for pointers on how to contribute to `gopy`.

## Documentation

A presentation was given at [dotgo-2015](http://talks.godoc.org/github.com/sbinet/talks/2015/20151109-gopy-dotgo/gopy-dotgo.slide).
A longer version of that talk is also available [here](http://talks.godoc.org/github.com/sbinet/talks/2015/20150929-gopy-lyon/gopy-lyon.slide#17).
An article was also posted on the [GopherAcademy Advent-2015](https://blog.gopheracademy.com/advent-2015/gopy/).

Documentation is available on [godoc](https://godoc.org):
 https://godoc.org/github.com/go-python/gopy

The `pkg` and `exe` commands are for end-users and create a full standalone python package that can be installed locally using `make install` based on the auto-generated `Makefile`.  Theoretically these packages could be uploaded to https://pypi.org/ for wider distribution, but that would require a lot more work to handle all the different possible python versions and coordination with the Go source version, so it is much better to just do the local make install on your system.  The `gen` and `build` commands are used for testing and just generate / build the raw binding files only.

IMPORTANT: many errors will be avoided by specifying the `-vm` option to gopy, with a full path if needed, or typically just `-vm=python3` to use python3 instead of version 2, which is often the default for the plain `python` command.

Here are some (slightly enhanced) docs from the help command:

```sh
$ gopy help
gopy - 

Commands:

    pkg         generate and compile Python bindings for Go, automatically including subdirs
                    also creates all the python files needed to install module
    exe         like pkg but makes a standalone executable with Go packages bultin
                    this is particularly useful when using -main arg to start process on
    gen         generate (C)Python language bindings for Go
    build       generate and compile 
                    main thread -- python interpreter can run on another thread.

Use "gopy help <command>" for more information about a command.


$ gopy help pkg
Usage: gopy pkg <go-package-name> [other-go-package...]

pkg generates and compiles (C)Python language bindings for a Go package, including subdirectories, and generates python module packaging suitable for distribution.  if setup.py file does not yet exist in the target directory, then it is created along with other default packaging files, using arguments.  Typically you create initial default versions of these files and then edit them, and after that, only regenerate the Go binding files.

ex:
 $ gopy pkg [options] <go-package-name> [other-go-package...]
 $ gopy pkg github.com/go-python/gopy/_examples/hi

Options:
  -author="gopy": author name
  -desc="": short description of project (long comes from README.md)
  -email="gopy@example.com": author email
  -exclude="": comma-separated list of package names to exclude
  -main="": code string to run in the Go GoPyInit() function in the cgo library
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for root of package
  -symbols=true: include symbols in output
  -url="https://github.com/go-python/gopy": home page for project
  -user="": username on https://www.pypa.io/en/latest/ for package name suffix
  -version="0.1.0": semantic version number -- can use e.g., git to get this from tag and pass as argument
  -vm="python": path to python interpreter


$ gopy help exe
Usage: gopy exe <go-package-name> [other-go-package...]

exe generates and compiles (C)Python language bindings for a Go package, including subdirectories, and generates a standalone python executable and associated module packaging suitable for distribution.  if setup.py file does not yet exist in the target directory, then it along with other default packaging files are created, using arguments.  Typically you create initial default versions of these files and then edit them, and after that, only regenerate the Go binding files.

The primary need for an exe instead of a pkg dynamic library is when the main thread must be used for something other than running the python interpreter, such as for a GUI library where the main thread must be used for running the GUI event loop (e.g., GoGi).

ex:
 $ gopy exe [options] <go-package-name> [other-go-package...]
 $ gopy exe github.com/go-python/gopy/_examples/hi

Options:
  -author="gopy": author name
  -desc="": short description of project (long comes from README.md)
  -email="gopy@example.com": author email
  -exclude="": comma-separated list of package names to exclude
  -main="": code string to run in the Go main() function in the cgo library -- defaults to GoPyMainRun() but typically should be overriden
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for root of package
  -symbols=true: include symbols in output
  -url="https://github.com/go-python/gopy": home page for project
  -user="": username on https://www.pypa.io/en/latest/ for package name suffix
  -version="0.1.0": semantic version number -- can use e.g., git to get this from tag and pass as argument
  -vm="python": path to python interpreter

$ gopy help gen
Usage: gopy gen <go-package-name> [other-go-package...]

gen generates (C)Python language bindings for Go package(s).

ex:
 $ gopy gen [options] <go-package-name> [other-go-package...]
 $ gopy gen github.com/go-python/gopy/_examples/hi

Options:
  -main="": code string to run in the Go main() function in the cgo library
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
  -main="": code string to run in the Go main() function in the cgo library
  -name="": name of output package (otherwise name of first package is used)
  -output="": output directory for bindings
  -symbols=true: include symbols in output
  -vm="python": path to python interpreter

```

## Examples

### From the command line

Note: you now need to make a go.mod file if you don't already have one in your environment, and get the package before building:

```sh
$ go mod init dummy.com/dum
$ go get github.com/go-python/gopy/_examples/hi
$ gopy build -output=out -vm=python3 github.com/go-python/gopy/_examples/hi
$ ls out
Makefile  __init__.py  __pycache__/  _hi.so*  build.py  go.py  hi.c  hi.go  hi.py  hi_go.h  hi_go.so
```

```sh
$ cd out
$ python3
>>> from out import hi
>>> dir(hi)
['Add', 'Concat', 'Hello', 'Hi', 'NewPerson', 'Person', '__doc__', '__file__', '__name__', '__package__']

>>> hi.Hello("you")
hello you from go

```

You can also run:

```sh
go test -v -run=TestHi
...
```

### From the `python` shell (NOT YET WORKING)

NOTE: following not yet working in new version:

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

## Troubleshooting

### python version mismatches

Many errors will be avoided by specifying the `-vm` option to gopy, with a full path if needed, or typically just `-vm=python3` to use python3 instead of version 2, which is often the default for the plain `python` command.

If you get any kind of error about the library module not being able to be imported, or apparently a large number of other random-looking errors, a mismatch between the python version used for compiling vs. what you are using to run is the most likely explanation.

If you get an error like this after importing a generated module:
```bash
Fatal Python error: _PyInterpreterState_Get(): no current thread state
```
it means you are running a different version of python than the one that build the library you are importing -- make sure you've got the paths in your `-vm` arg aligned with what you are using to import.

### linux: cannot find .so file

If your `import` statement fails to find the module `.so` file, and it is in the current directory, you may need to ensure that the linker `ld` will look in the current directory for library files -- add this to your `.bashrc` file (and `source` that file after editing, or enter command locally):

```sh
export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.
```

you may also need to add your python path, but that is less likely to be an issue.

## Contribute

`gopy` is part of the `go-python` organization and licensed under `BSD-3`.
When you want to contribute a patch or some code to `gopy`, please send a pull
request against the `gopy` issue tracker **AND** a pull request against
[go-python/license](https://github.com/go-python/license) adding yourself to the
`AUTHORS` and `CONTRIBUTORS` files.

