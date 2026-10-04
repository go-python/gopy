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

	if err := runCmd(nil, "goimports", "-w", cfg.Name+".go"); err != nil {
		return err
	}

	if cfg.Backend == bind.BackendCFFI {
		return buildCFFI(cfg, buildname+libExt)
	}

	pycfg, err := bind.GetPythonConfig(cfg.VM)
	if err != nil {
		return err
	}

	switch cfg.Backend {
	case bind.BackendPyBind11:
		return buildPyBind11(cfg, pycfg)
	case bind.BackendNanobind:
		return buildNanobind(cfg, pycfg)
	}

	if mode == bind.ModeExe {
		of, err := os.Create(buildname + ".h") // overwrite existing
		fmt.Fprintf(of, "#if !defined(__STDC_VERSION__) || (__STDC_VERSION__ < 202311L)\ntypedef uint8_t bool;\n#endif\n")
		of.Close()

		fmt.Printf("%v build.py   # will fail, but needed to generate .c file\n", cfg.VM)
		cmd := exec.Command(cfg.VM, "build.py")
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
		if err := runCmd(nil, "go", goBuildArgs(cfg, "c-shared", buildLib)...); err != nil {
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
		var exportLdFlags []string
		switch runtime.GOOS {
		case "darwin":
			ef, ferr := os.CreateTemp("", "gopy-exports-*.txt")
			if ferr == nil {
				fmt.Fprintf(ef, "_PyInit__%s\n", cfg.Name)
				ef.Close()
				defer os.Remove(ef.Name())
				exportLdFlags = append(exportLdFlags, "-extldflags=-Wl,-exported_symbols_list,"+ef.Name())
			}
		case "linux":
			ef, ferr := os.CreateTemp("", "gopy-exports-*.map")
			if ferr == nil {
				fmt.Fprintf(ef, "{ global: PyInit__%s; local: *; };\n", cfg.Name)
				ef.Close()
				defer os.Remove(ef.Name())
				exportLdFlags = append(exportLdFlags, "-extldflags=-Wl,--version-script="+ef.Name())
			}
		}

		// generate c code
		if err := runCmd(nil, cfg.VM, "build.py"); err != nil {
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
		if err := runCmd(env, "go", goBuildArgs(cfg, "c-shared", modlib, exportLdFlags...)...); err != nil {
			return err
		}
	}

	return err
}

// buildCFFI builds the cgo shim as a plain shared library, and then runs
// build.py to write the cffi module that loads it.  The current directory
// is the output directory.
func buildCFFI(cfg *BuildCfg, buildLib string) error {
	if err := runCmd(nil, "go", goBuildArgs(cfg, "c-shared", buildLib)...); err != nil {
		return err
	}
	return runCmd(nil, cfg.VM, "build.py")
}

// buildPyBind11 builds the pybind11 backend's module (see buildCXXModule).
func buildPyBind11(cfg *BuildCfg, pycfg bind.PyConfig) error {
	cmdout, err := exec.Command(cfg.VM, "-m", "pybind11", "--includes").CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n(is pybind11 installed? pip install pybind11)\n", err, string(cmdout))
		return err
	}
	return buildCXXModule(cfg, pycfg, strings.Fields(strings.TrimSpace(string(cmdout))), nil)
}

// buildNanobind builds the nanobind backend's module (see buildCXXModule).
// Unlike pybind11, nanobind isn't header-only: its own runtime (libnanobind)
// ships as source, meant to be compiled into each extension alongside the
// extension's own code, which nb_combined.cpp does in one translation unit.
func buildNanobind(cfg *BuildCfg, pycfg bind.PyConfig) error {
	cmdout, err := exec.Command(cfg.VM, "-c",
		"import nanobind; print(nanobind.include_dir()); print(nanobind.source_dir())").CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n(is nanobind installed? pip install nanobind)\n", err, string(cmdout))
		return err
	}
	dirs := strings.Split(strings.TrimSpace(string(cmdout)), "\n")
	if len(dirs) != 2 {
		return fmt.Errorf("gopy: unexpected output locating nanobind: %q", string(cmdout))
	}
	incdir, srcdir := strings.TrimSpace(dirs[0]), strings.TrimSpace(dirs[1])
	// robin_map is a dependency nanobind vendors next to its own headers.
	robinmap := filepath.Join(filepath.Dir(incdir), "ext", "robin_map", "include")
	flags := append([]string{"-I" + incdir, "-I" + robinmap}, bind.NanobindCXXFlags...)
	return buildCXXModule(cfg, pycfg, flags, []string{filepath.Join(srcdir, "nb_combined.cpp")})
}

// buildCXXModule builds the cgo shim as a static archive, runs build.py to
// write a C++ module wrapping it (pybind11 or nanobind), and compiles+links
// that with a C++ compiler (see bind.CXXArgs, and bind/cxxbuild.go for why
// a static archive), passing it cxxflags (the C++ library's include
// directories, and any flags of its own) and, besides the generated .cpp,
// the C++ library's own sources, if any.  The current directory is the
// output directory.  The Makefile gopy gen writes for these backends runs
// the same steps.
func buildCXXModule(cfg *BuildCfg, pycfg bind.PyConfig, cxxflags, srcs []string) error {
	if err := runCmd(nil, "go", goBuildArgs(cfg, "c-archive", bind.CXXArchive(cfg.Name))...); err != nil {
		return err
	}
	if err := runCmd(nil, cfg.VM, "build.py"); err != nil {
		return err
	}
	cxxArgs := bind.CXXArgs(cfg.Name, bind.ExtModuleName(cfg.Name, libExt, pycfg), pycfg, cxxflags, srcs)
	return runCmd(nil, bind.CXX(), cxxArgs...)
}

// goBuildArgs returns the go build arguments that build the package in the
// current directory into out with the given -buildmode, with cfg's build
// tags and, unless cfg.Symbols is set, without symbol tables (-s omits the
// symbol table and debug information, -w the DWARF symbol table; see
// https://golang.org/cmd/link/).  ldflags are any further linker flags.
func goBuildArgs(cfg *BuildCfg, buildmode, out string, ldflags ...string) []string {
	args := []string{"build", "-mod=mod", "-buildmode=" + buildmode}
	if cfg.BuildTags != "" {
		args = append(args, "-tags", cfg.BuildTags)
	}
	if !cfg.Symbols {
		ldflags = append([]string{"-s", "-w"}, ldflags...)
	}
	if len(ldflags) > 0 {
		args = append(args, "-ldflags="+strings.Join(ldflags, " "))
	}
	return append(args, "-o", out, ".")
}

// runCmd runs name with args, in env if it isn't nil, printing the command
// line first, and its output if it fails.
func runCmd(env []string, name string, args ...string) error {
	fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(out))
	}
	return err
}
