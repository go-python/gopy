// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"

	"github.com/go-python/gopy/bind"
)

func gopyMakeCmdBuild() *commander.Command {
	cmd := &commander.Command{
		Run:       gopyRunCmdBuild,
		UsageLine: "build <go-package-name> [other-go-package...]",
		Short:     "generate and compile (C)Python language bindings for Go",
		Long: `
build generates and compiles (C)Python language bindings for Go package(s).

ex:
 $ gopy build [options] <go-package-name> [other-go-package...]
 $ gopy build github.com/go-python/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-build", flag.ExitOnError),
	}

	cmd.Flag.String("vm", "python", "path to python interpreter")
	cmd.Flag.String("output", "", "output directory for bindings")
	cmd.Flag.String("name", "", "name of output package (otherwise name of first package is used)")
	cmd.Flag.String("main", "", "code string to run in the go main() function in the cgo library")
	cmd.Flag.String("package-prefix", ".", "custom package prefix used when generating import "+
		"statements for generated package")
	cmd.Flag.Bool("rename", false, "rename Go symbols to python PEP snake_case")
	cmd.Flag.Bool("symbols", true, "include symbols in output")
	cmd.Flag.Bool("no-warn", false, "suppress warning messages, which may be expected")
	cmd.Flag.Bool("no-make", false, "do not generate a Makefile, e.g., when called from Makefile")
	cmd.Flag.Bool("clear-go-tls", false, "emit a _gopy_clear_go_tls() call before every CGo entry (issue #370); off by default, needed only when several gopy extensions share one process and known to crash CPython 3.12+ (issue #395)")
	cmd.Flag.Bool("dynamic-link", false, "whether to link output shared library dynamically to Python")
	cmd.Flag.String("build-tags", "", "build tags to be passed to `go build`")
	return cmd
}

func gopyRunCmdBuild(cmdr *commander.Command, args []string) error {
	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cfg := NewBuildCfg()
	cfg.OutputDir = cmdr.Flag.Lookup("output").Value.Get().(string)
	cfg.Name = cmdr.Flag.Lookup("name").Value.Get().(string)
	cfg.Main = cmdr.Flag.Lookup("main").Value.Get().(string)
	cfg.VM = cmdr.Flag.Lookup("vm").Value.Get().(string)
	cfg.PkgPrefix = cmdr.Flag.Lookup("package-prefix").Value.Get().(string)
	cfg.RenameCase = cmdr.Flag.Lookup("rename").Value.Get().(bool)
	cfg.Symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)
	cfg.NoWarn = cmdr.Flag.Lookup("no-warn").Value.Get().(bool)
	cfg.NoMake = cmdr.Flag.Lookup("no-make").Value.Get().(bool)
	cfg.DynamicLinking = cmdr.Flag.Lookup("dynamic-link").Value.Get().(bool)
	cfg.BuildTags = cmdr.Flag.Lookup("build-tags").Value.Get().(string)

	bind.NoWarn = cfg.NoWarn
	bind.NoMake = cfg.NoMake
	bind.ClearGoTLS = cmdr.Flag.Lookup("clear-go-tls").Value.Get().(bool)

	for _, path := range args {
		bpkg, err := loadPackage(path, true, cfg.BuildTags) // build first
		if err != nil {
			return fmt.Errorf("gopy-gen: go build / load of package failed with path=%q: %v", path, err)
		}
		pkg, err := parsePackage(bpkg)
		if err != nil {
			return err
		}
		if cfg.Name == "" {
			cfg.Name = pkg.Name()
		}
	}
	return runBuild("build", cfg)
}

// runBuild calls genPkg and then executes commands to build the resulting files
// exe = executable mode to build an executable instead of a library
// mode = gen, build, pkg, exe
func runBuild(mode bind.BuildMode, cfg *BuildCfg) error {
	var err error
	cfg.OutputDir, err = genOutDir(cfg.OutputDir)
	if err != nil {
		return err
	}
	err = genPkg(mode, cfg)
	if err != nil {
		return err
	}

	fmt.Printf("\n--- building package ---\n%s\n", cfg.Cmd)

	buildname := cfg.Name + "_go"
	var cmdout []byte
	cwd, err := os.Getwd()
	os.Chdir(cfg.OutputDir)
	defer os.Chdir(cwd)

	os.Remove(cfg.Name + ".c") // may fail, we don't care

	fmt.Printf("goimports -w %v\n", cfg.Name+".go")
	cmd := exec.Command("goimports", "-w", cfg.Name+".go")
	cmdout, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\no%v\n", err, string(cmdout))
		return err
	}

	if cfg.Backend == bind.BackendCFFI {
		return buildCFFI(cfg, buildname+libExt)
	}

	pycfg, err := bind.GetPythonConfig(cfg.VM)
	if err != nil {
		return err
	}

	if cfg.Backend == bind.BackendPyBind11 {
		return buildPyBind11(cfg, buildname+libExt, pycfg)
	}

	if mode == bind.ModeExe {
		of, err := os.Create(buildname + ".h") // overwrite existing
		fmt.Fprintf(of, "#if !defined(__STDC_VERSION__) || (__STDC_VERSION__ < 202311L)\ntypedef uint8_t bool;\n#endif\n")
		of.Close()

		fmt.Printf("%v build.py   # will fail, but needed to generate .c file\n", cfg.VM)
		cmd = exec.Command(cfg.VM, "build.py")
		cmd.Run() // will fail, we don't care about errors

		args := []string{"build", "-mod=mod", "-buildmode=c-shared"}
		if cfg.BuildTags != "" {
			args = append(args, "-tags", cfg.BuildTags)
		}
		args = append(args, "-o", buildname+libExt, ".")

		fmt.Printf("go %v\n", strings.Join(args, " "))
		cmd = exec.Command("go", args...)
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
			return err
		}

		fmt.Printf("%v build.py   # should work this time\n", cfg.VM)
		cmd = exec.Command(cfg.VM, "build.py")
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
			return err
		}

		err = os.Remove(cfg.Name + "_go" + libExt)

		fmt.Printf("go build -o py%s\n", cfg.Name)
		cmd = exec.Command("go", "build", "-mod=mod")
		if cfg.BuildTags != "" {
			args = append(args, "-tags", cfg.BuildTags)
		}
		args = append(args, "-o", "py"+cfg.Name)
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
			return err
		}

	} else {
		buildLib := buildname + libExt
		extext := libExt
		if runtime.GOOS == "windows" {
			extext = ".pyd"
		}
		if pycfg.ExtSuffix != "" {
			extext = pycfg.ExtSuffix
		}
		modlib := "_" + cfg.Name + extext

		// build the go shared library upfront to generate the header
		// needed by our generated cpython code
		firstArgs := []string{"build", "-mod=mod", "-buildmode=c-shared"}
		if cfg.BuildTags != "" {
			firstArgs = append(firstArgs, "-tags", cfg.BuildTags)
		}
		if !cfg.Symbols {
			// These flags will omit the various symbol tables, thereby
			// reducing the final size of the binary. From https://golang.org/cmd/link/
			// -s Omit the symbol table and debug information
			// -w Omit the DWARF symbol table
			firstArgs = append(firstArgs, "-ldflags=-s -w")
		}
		firstArgs = append(firstArgs, "-o", buildLib, ".")
		fmt.Printf("go %v\n", strings.Join(firstArgs, " "))
		cmd = exec.Command("go", firstArgs...)
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
			return err
		}
		// we don't need this initial lib because we are going to relink
		os.Remove(buildLib)

		// Build the final extension with symbol-visibility restriction so that
		// Go runtime globals are not placed in the global dynamic-linker
		// namespace. Two independently-loaded Go runtimes sharing those globals
		// via RTLD_GLOBAL interposition corrupt each other's GC state (#370).
		// This applies only to the second build, which is where PyInit__<name>
		// exists and where the exported-symbols list is valid.
		finalArgs := []string{"build", "-mod=mod", "-buildmode=c-shared"}
		if cfg.BuildTags != "" {
			finalArgs = append(finalArgs, "-tags", cfg.BuildTags)
		}
		var finalLdFlags []string
		if !cfg.Symbols {
			finalLdFlags = append(finalLdFlags, "-s", "-w")
		}
		switch runtime.GOOS {
		case "darwin":
			ef, ferr := os.CreateTemp("", "gopy-exports-*.txt")
			if ferr == nil {
				fmt.Fprintf(ef, "_PyInit__%s\n", cfg.Name)
				ef.Close()
				defer os.Remove(ef.Name())
				finalLdFlags = append(finalLdFlags, "-extldflags=-Wl,-exported_symbols_list,"+ef.Name())
			}
		case "linux":
			ef, ferr := os.CreateTemp("", "gopy-exports-*.map")
			if ferr == nil {
				fmt.Fprintf(ef, "{ global: PyInit__%s; local: *; };\n", cfg.Name)
				ef.Close()
				defer os.Remove(ef.Name())
				finalLdFlags = append(finalLdFlags, "-extldflags=-Wl,--version-script="+ef.Name())
			}
		}
		if len(finalLdFlags) > 0 {
			finalArgs = append(finalArgs, "-ldflags="+strings.Join(finalLdFlags, " "))
		}
		finalArgs = append(finalArgs, "-o", modlib, ".")
		// args is still used below for the CGO env build; point it at finalArgs.
		args := finalArgs

		// generate c code
		fmt.Printf("%v build.py\n", cfg.VM)
		cmd = exec.Command(cfg.VM, "build.py")
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\no%v\n", err, string(cmdout))
			return err
		}

		if bind.WindowsOS {
			fmt.Printf("Doing windows sed hack to fix declspec for PyInit\n")
			fname := cfg.Name + ".c"
			raw, err := os.ReadFile(fname)
			if err != nil {
				fmt.Printf("could not read %s: %+v", fname, err)
				return fmt.Errorf("could not read %s: %w", fname, err)
			}
			raw = bytes.ReplaceAll(raw, []byte(" PyInit_"), []byte(" __declspec(dllexport) PyInit_"))
			err = os.WriteFile(fname, raw, 0644)
			if err != nil {
				fmt.Printf("could not apply sed hack to fix declspec for PyInit: %+v", err)
				return fmt.Errorf("could not apply sed hack to fix PyInit: %w", err)
			}
		}

		cflags := strings.Fields(strings.TrimSpace(pycfg.CFlags))
		cflags = append(cflags, "-fPIC", "-O3", "-ffast-math")
		if include, exists := os.LookupEnv("GOPY_INCLUDE"); exists {
			cflags = append(cflags, "-I"+filepath.ToSlash(include))
		}
		if oldcflags, exists := os.LookupEnv("CGO_CFLAGS"); exists {
			cflags = append(cflags, oldcflags)
		}
		var ldflags []string
		if cfg.DynamicLinking {
			ldflags = strings.Fields(strings.TrimSpace(pycfg.LdDynamicFlags))
		} else {
			ldflags = strings.Fields(strings.TrimSpace(pycfg.LdFlags))
		}
		if !cfg.Symbols {
			ldflags = append(ldflags, "-s")
		}
		if lib, exists := os.LookupEnv("GOPY_LIBDIR"); exists {
			ldflags = append(ldflags, "-L"+filepath.ToSlash(lib))
		}
		if libname, exists := os.LookupEnv("GOPY_PYLIB"); exists {
			ldflags = append(ldflags, "-l"+filepath.ToSlash(libname))
		}
		if oldldflags, exists := os.LookupEnv("CGO_LDFLAGS"); exists {
			ldflags = append(ldflags, oldldflags)
		}

		removeEmpty := func(src []string) []string {
			o := make([]string, 0, len(src))
			for _, v := range src {
				if v == "" {
					continue
				}
				o = append(o, v)
			}
			return o
		}

		cflags = removeEmpty(cflags)
		ldflags = removeEmpty(ldflags)

		cflagsEnv := fmt.Sprintf("CGO_CFLAGS=%s", strings.Join(cflags, " "))
		ldflagsEnv := fmt.Sprintf("CGO_LDFLAGS=%s", strings.Join(ldflags, " "))

		env := os.Environ()
		env = append(env, cflagsEnv)
		env = append(env, ldflagsEnv)

		fmt.Println(cflagsEnv)
		fmt.Println(ldflagsEnv)

		// build extension with go + c
		fmt.Printf("go %v\n", strings.Join(args, " "))
		cmd = exec.Command("go", args...)
		cmd.Env = env
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
			return err
		}
	}

	return err
}

// buildCFFI builds the cgo shim as a plain shared library, and then runs
// build.py to write the cffi module that loads it.  The current directory
// is the output directory.
func buildCFFI(cfg *BuildCfg, buildLib string) error {
	args := []string{"build", "-mod=mod", "-buildmode=c-shared"}
	if cfg.BuildTags != "" {
		args = append(args, "-tags", cfg.BuildTags)
	}
	if !cfg.Symbols {
		args = append(args, "-ldflags=-s -w")
	}
	args = append(args, "-o", buildLib, ".")
	fmt.Printf("go %v\n", strings.Join(args, " "))
	cmdout, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
		return err
	}

	fmt.Printf("%v build.py\n", cfg.VM)
	cmdout, err = exec.Command(cfg.VM, "build.py").CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
	}
	return err
}

// buildPyBind11 builds the cgo shim as a plain shared library (same shape as
// buildCFFI's), runs build.py to write a pybind11 C++ module wrapping it, and
// compiles+links that with a C++ compiler.  The current directory is the
// output directory.
func buildPyBind11(cfg *BuildCfg, buildLib string, pycfg bind.PyConfig) error {
	args := []string{"build", "-mod=mod", "-buildmode=c-shared"}
	if cfg.BuildTags != "" {
		args = append(args, "-tags", cfg.BuildTags)
	}
	if !cfg.Symbols {
		args = append(args, "-ldflags=-s -w")
	}
	args = append(args, "-o", buildLib, ".")
	fmt.Printf("go %v\n", strings.Join(args, " "))
	cmdout, err := exec.Command("go", args...).CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
		return err
	}

	fmt.Printf("%v build.py\n", cfg.VM)
	cmdout, err = exec.Command(cfg.VM, "build.py").CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
		return err
	}

	cmdout, err = exec.Command(cfg.VM, "-m", "pybind11", "--includes").CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n(is pybind11 installed? pip install pybind11)\n", err, string(cmdout))
		return err
	}
	pyinc := strings.Fields(strings.TrimSpace(string(cmdout)))

	extext := libExt
	if runtime.GOOS == "windows" {
		extext = ".pyd"
	}
	if pycfg.ExtSuffix != "" {
		extext = pycfg.ExtSuffix
	}
	modlib := "_" + cfg.Name + extext

	cxx := os.Getenv("CXX")
	if cxx == "" {
		cxx = "c++"
	}
	// pycfg.CFlags/LdFlags quote each path (for the shell that CGO_CFLAGS/
	// CGO_LDFLAGS normally go through); exec.Command runs the compiler
	// directly, with no shell to strip those, so unquote each field here.
	unquote := func(fields []string) []string {
		o := make([]string, len(fields))
		for i, f := range fields {
			o[i] = strings.Trim(f, `"`)
		}
		return o
	}
	// modlib depends on buildLib (alongside it) and libpython (wherever this
	// VM's own one lives, e.g. not on the loader's default search path for a
	// uv- or pyenv-managed Python); without an rpath for each, the loader
	// only finds them if they happen to already be on its search path.
	var libdir string
	if m := regexp.MustCompile(`-L(\S+)`).FindStringSubmatch(pycfg.LdFlags); m != nil {
		libdir = strings.Trim(m[1], `"`)
	}
	cxxArgs := []string{"-std=c++17", "-fPIC", "-shared", "-O2"}
	switch runtime.GOOS {
	case "darwin":
		cxxArgs = append(cxxArgs, "-Wl,-rpath,@loader_path")
		if libdir != "" {
			cxxArgs = append(cxxArgs, "-Wl,-rpath,"+libdir)
		}
	case "windows":
		// No rpath equivalent, but modlib finding buildLib (in the same
		// directory) is handled at import time instead, by the generated
		// wrapper's os.add_dll_directory() call (see PyWrapPreamble).
		// MinGW's own runtime (libstdc++/libgcc/libwinpthread), which g++
		// links dynamically by default, has no such fix available -- it
		// isn't found by name alone unless its directory happens to be on
		// PATH -- so link it in statically instead.  The C runtime (ucrt)
		// stays dynamic, shared with Python's own.
		cxxArgs = append(cxxArgs, "-static-libgcc", "-static-libstdc++",
			"-Wl,-Bstatic,--whole-archive", "-lwinpthread", "-Wl,--no-whole-archive", "-Wl,-Bdynamic")
	default:
		cxxArgs = append(cxxArgs, "-Wl,-rpath,$ORIGIN")
		if libdir != "" {
			cxxArgs = append(cxxArgs, "-Wl,-rpath,"+libdir)
		}
	}
	cxxArgs = append(cxxArgs, pyinc...)
	cxxArgs = append(cxxArgs, unquote(strings.Fields(pycfg.CFlags))...)
	cxxArgs = append(cxxArgs, cfg.Name+".cpp", buildLib)
	cxxArgs = append(cxxArgs, unquote(strings.Fields(pycfg.LdFlags))...)
	cxxArgs = append(cxxArgs, "-o", modlib)
	fmt.Printf("%v %v\n", cxx, strings.Join(cxxArgs, " "))
	cmdout, err = exec.Command(cxx, cxxArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
	}
	return err
}
