// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bind

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// How the C++ backends (pybind11, nanobind) compile and link their module,
// shared by gopy build (buildCXXModule in cmd_build.go, which runs these
// steps itself) and the Makefile written by gopy gen/pkg (genMakefileCXX).
//
// The Go side builds as a static archive (-buildmode=c-archive), not a
// shared library: the generated .cpp defines the callback trampolines Go
// calls into (see pybind11_callback.go) as well as calling into Go itself,
// and two separately-built shared libraries can't have a dependency cycle
// like that -- neither can exist as a complete, loadable file before the
// other -- so both sides' symbols are resolved in the single final link
// that CXXArgs describes.

// NanobindCXXFlags are the flags nanobind's own build uses for libnanobind
// (see the comment at the top of its nb_combined.cpp), which the nanobind
// backend compiles into every module; harmless for the generated .cpp too.
var NanobindCXXFlags = []string{"-DNDEBUG", "-DNB_COMPACT_ASSERTIONS", "-fno-strict-aliasing"}

// CXX returns the C++ compiler to use: $CXX, or else c++.
func CXX() string {
	if cxx := os.Getenv("CXX"); cxx != "" {
		return cxx
	}
	return "c++"
}

// CXXArchive returns the file name of the static archive the cgo shim for
// package name builds as.
func CXXArchive(name string) string {
	return name + "_go.a"
}

// ExtModuleName returns the file name of the extension module _<name>,
// preferring the interpreter's own suffix (e.g. .cpython-312-x86_64-linux-gnu.so)
// over libext.
func ExtModuleName(name, libext string, pycfg PyConfig) string {
	if pycfg.ExtSuffix != "" {
		return "_" + name + pycfg.ExtSuffix
	}
	return "_" + name + libext
}

// CXXArgs returns the C++ compiler arguments that compile <name>.cpp and
// srcs (the C++ library's own sources, if any) with libflags (its include
// directories and flags of its own), and link them with the archive into
// modlib.
func CXXArgs(name, modlib string, pycfg PyConfig, libflags, srcs []string) []string {
	archive := CXXArchive(name)
	// pycfg.CFlags/LdFlags quote each path (for the shell that CGO_CFLAGS/
	// CGO_LDFLAGS normally go through); these arguments go to the compiler
	// directly, so unquote each field here.
	unquote := func(fields []string) []string {
		o := make([]string, len(fields))
		for i, f := range fields {
			o[i] = strings.Trim(f, `"`)
		}
		return o
	}
	// modlib depends on libpython (wherever this VM's own one lives, e.g.
	// not on the loader's default search path for a uv- or pyenv-managed
	// Python); without an rpath, the loader only finds it if it happens to
	// already be on its search path.
	var libdir string
	if m := regexp.MustCompile(`-L(\S+)`).FindStringSubmatch(pycfg.LdFlags); m != nil {
		libdir = strings.Trim(m[1], `"`)
	}
	// The archive's Go runtime code calls into gopy_cb_N (defined in the
	// .cpp), so the linker must be told to keep every object in it -- left
	// to its own judgement, it would see nothing in the .cpp calling into
	// the archive first and drop it as unused.  GNU ld (Linux, and Windows'
	// MinGW) and ld64 (macOS) spell that differently.
	var archiveArgs []string
	if runtime.GOOS == "darwin" {
		archiveArgs = []string{"-Wl,-force_load," + archive}
	} else {
		archiveArgs = []string{"-Wl,--whole-archive", archive, "-Wl,--no-whole-archive"}
	}
	args := []string{"-std=c++17", "-fPIC", "-shared", "-O2"}
	switch runtime.GOOS {
	case "darwin":
		args = append(args, "-Wl,-rpath,@loader_path")
		if libdir != "" {
			args = append(args, "-Wl,-rpath,"+libdir)
		}
	case "windows":
		// No rpath equivalent; modlib depends on nothing but libpython, the
		// Go side being a static archive rather than a separate DLL of its
		// own.  MinGW's own runtime (libstdc++/libgcc/libwinpthread), which
		// g++ links dynamically by default, has no such fix available -- it
		// isn't found by name alone unless its directory happens to be on
		// PATH -- so link it in statically instead.  The C runtime (ucrt)
		// stays dynamic, shared with Python's own.
		args = append(args, "-static-libgcc", "-static-libstdc++",
			"-Wl,-Bstatic,--whole-archive", "-lwinpthread", "-Wl,--no-whole-archive", "-Wl,-Bdynamic")
	default:
		args = append(args, "-Wl,-rpath,$ORIGIN")
		if libdir != "" {
			args = append(args, "-Wl,-rpath,"+libdir)
		}
	}
	args = append(args, libflags...)
	args = append(args, unquote(strings.Fields(pycfg.CFlags))...)
	args = append(args, name+".cpp")
	args = append(args, srcs...)
	args = append(args, archiveArgs...)
	args = append(args, unquote(strings.Fields(pycfg.LdFlags))...)
	// c-archive mode (unlike c-shared) doesn't resolve the Go runtime's own
	// dependencies on these itself; TODO: verified only on Linux -- unclear
	// yet whether Windows/macOS need anything of their own added here too.
	if runtime.GOOS != "windows" {
		args = append(args, "-lpthread", "-ldl", "-lm")
	}
	return append(args, "-o", modlib)
}

// makeShellArg returns arg written into a Makefile recipe line, so that the
// shell make runs it with receives arg itself: make variable references
// ("$(...)") are left for make to expand, any other "$" is escaped from
// make, and anything the shell would split or expand is single-quoted.
// Each "'" inside closes the quoting, adds a double-quoted "'", and reopens
// it, rather than using a backslash: the double quote makes make hand the
// line to the shell, where Windows make's own argument splitting would
// otherwise drop the "'".
func makeShellArg(arg string) string {
	if strings.HasPrefix(arg, "$(") {
		return arg
	}
	arg = strings.ReplaceAll(arg, "$", "$$")
	if arg == "" || strings.ContainsAny(arg, " \t\n'\"\\`*?[#~&;|<>()$") {
		return "'" + strings.ReplaceAll(arg, "'", `'"'"'`) + "'"
	}
	return arg
}

// MakefileTemplateCXX is the Makefile for the C++ backends: 1 = package
// name, 2 = gopy command, 3 = gencmd, 4 = vm, 5 = C++ compiler, 6 = the
// backend's own make variables (CXXLIBFOUND, CXXLIBFLAGS, CXXLIBSRCS), 7 = C++ compiler
// arguments, 8 = gopy version, 9 = backend name, 10 = module file name.
const MakefileTemplateCXX = `# Makefile for python interface for package %[1]s, using %[9]s.
# File is generated by gopy version %[8]s. Do not edit.
# %[2]s

GOCMD=go
GOBUILD=$(GOCMD) build -mod=mod
GOIMPORTS=goimports
PYTHON=%[4]s
CXX=%[5]s
%[6]s
all: gen build

gen:
	%[3]s

build:
	# $(shell ...) expands to nothing, rather than failing, if $(PYTHON) can't find %[9]s
	@test -n "$(CXXLIBFOUND)" || { echo "%[9]s not found for $(PYTHON) (pip install %[9]s)" >&2; exit 1; }
	# goimports is needed to ensure that the imports list is valid
	$(GOIMPORTS) -w %[1]s.go
	# build %[1]s_go.a from %[1]s.go -- the cgo wrappers to go functions -- as a
	# static archive: the module below links it in, rather than loading it
	$(GOBUILD) -buildmode=c-archive -o %[1]s_go.a %[1]s.go
	# writes %[1]s.cpp, the %[9]s module wrapping it
	$(PYTHON) build.py
	# compile and link %[10]s, the module %[1]s.py imports
	$(CXX) %[7]s

`

// genMakefileCXX writes the Makefile for the C++ backends.
func (g *pyGen) genMakefileCXX(gencmd string, pycfg PyConfig) {
	// The C++ library's own include directories and sources are looked up
	// when make runs, not now, so that gopy gen works without it installed.
	var libvars string
	switch {
	case g.isPyBind11():
		libvars = `CXXLIBFLAGS=$(shell $(PYTHON) -m pybind11 --includes)
CXXLIBSRCS=
CXXLIBFOUND=$(CXXLIBFLAGS)
`
	case g.isNanobind():
		libvars = `NANOBIND_INC=$(shell $(PYTHON) -c "import nanobind; print(nanobind.include_dir())")
NANOBIND_SRC=$(shell $(PYTHON) -c "import nanobind; print(nanobind.source_dir())")
# robin_map is a dependency nanobind vendors next to its own headers
CXXLIBFLAGS=-I$(NANOBIND_INC) -I$(NANOBIND_INC)/../ext/robin_map/include ` + strings.Join(NanobindCXXFlags, " ") + `
CXXLIBSRCS=$(NANOBIND_SRC)/nb_combined.cpp
CXXLIBFOUND=$(NANOBIND_INC)
`
	}
	modlib := ExtModuleName(g.cfg.Name, g.libext, pycfg)
	args := CXXArgs(g.cfg.Name, modlib, pycfg, []string{"$(CXXLIBFLAGS)"}, []string{"$(CXXLIBSRCS)"})
	for i, a := range args {
		args[i] = makeShellArg(a)
	}
	// make runs a $(shell ...) or recipe line through sh rather than
	// directly whenever it has quotes or other shell syntax in it (as the
	// nanobind lookups above do), and sh would strip a Windows path's
	// backslashes from $(PYTHON); forward slashes work either way.
	vm := filepath.ToSlash(g.cfg.VM)
	g.makefile.Printf(MakefileTemplateCXX, g.cfg.Name, g.cfg.Cmd, gencmd, vm, CXX(), libvars,
		strings.Join(args, " "), g.cfg.Version, g.cfg.Backend, modlib)
}
