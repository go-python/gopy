// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"

	"github.com/rudderlabs/gopy/bind"
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
 $ gopy build github.com/rudderlabs/gopy/_examples/hi
`,
		Flag: *flag.NewFlagSet("gopy-build", flag.ExitOnError),
	}

	AddCommonCmdFlags(&cmd.Flag)

	// Build Specific flags.
	cmd.Flag.String("package-prefix", ".", "custom package prefix used when generating import "+
		"statements for generated package")
	cmd.Flag.Bool("symbols", true, "include symbols in output")

	return cmd
}

func gopyRunCmdBuild(cmdr *commander.Command, args []string) error {
	if len(args) == 0 {
		err := fmt.Errorf("gopy: expect a fully qualified go package name as argument")
		log.Println(err)
		return err
	}

	cfg := NewBuildCfg(&cmdr.Flag)
	cfg.Symbols = cmdr.Flag.Lookup("symbols").Value.Get().(bool)

	bind.NoWarn = cfg.NoWarn
	bind.NoMake = cfg.NoMake
	bind.NoPyExceptions = cfg.NoPyExceptions

	for _, path := range args {
		bpkg, err := loadPackage(path, true) // build first
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

	if mode == bind.ModeExe {
		of, err := os.Create(buildname + ".h") // overwrite existing
		fmt.Fprintf(of, "typedef uint8_t bool;\n")
		of.Close()

		fmt.Printf("%v build.py   # will fail, but needed to generate .c file\n", cfg.VM)
		cmd = exec.Command(cfg.VM, "build.py")
		cmd.Run() // will fail, we don't care about errors

		args := []string{"build", "-mod=mod", "-buildmode=c-shared", "-o", buildname + libExt, "."}
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
		cmd = exec.Command("go", "build", "-mod=mod", "-o", "py"+cfg.Name)
		cmdout, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
			return err
		}

	} else {

		// build extension with go + c
		args, env, err := getBuildArgsAndEnv(cfg)
		if err != nil {
			fmt.Printf("error building environment: %v\n", err)
			return err
		}

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
			cmd = exec.Command("sed", "-i", "s/ PyInit_/ __declspec(dllexport) PyInit_/g", cfg.Name+".c")
			cmdout, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("cmd had error: %v  output:\no%v\n", err, string(cmdout))
				return err
			}
		}

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

func getBuildArgsAndEnv(cfg *BuildCfg) (args []string, env []string, err error) {
	buildname := cfg.Name + "_go"
	buildLib := buildname + libExt

	var pycfg bind.PyConfig
	pycfg, err = bind.GetPythonConfig(cfg.VM)
	if err != nil {
		fmt.Printf("error creating pycfg: %r\n", err)
		return args, env, err
	}

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
	args = []string{"build", "-mod=mod", "-buildmode=c-shared"}
	if !cfg.Symbols {
		// These flags will omit the various symbol tables, thereby
		// reducing the final size of the binary. From https://golang.org/cmd/link/
		// -s Omit the symbol table and debug information
		// -w Omit the DWARF symbol table
		args = append(args, "-ldflags=-s -w")
	}
	args = append(args, "-o", buildLib, ".")
	fmt.Printf("go %v\n", strings.Join(args, " "))
	cmd := exec.Command("go", args...)
	cmdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd had error: %v  output:\n%v\n", err, string(cmdout))
		return args, env, err
	}
	// update the output name to the one with the ABI extension
	args[len(args)-2] = modlib
	// we don't need this initial lib because we are going to relink
	os.Remove(buildLib)

	cflags := strings.Fields(strings.TrimSpace(pycfg.CFlags))
	cflags = append(cflags, "-fPIC", "-Ofast")
	if include, exists := os.LookupEnv("GOPY_INCLUDE"); exists {
		cflags = append(cflags, "-I"+filepath.ToSlash(include))
	}

	ldflags := strings.Fields(strings.TrimSpace(pycfg.LdFlags))
	if !cfg.Symbols {
		ldflags = append(ldflags, "-s")
	}
	if lib, exists := os.LookupEnv("GOPY_LIBDIR"); exists {
		ldflags = append(ldflags, "-L"+filepath.ToSlash(lib))
	}
	if libname, exists := os.LookupEnv("GOPY_PYLIB"); exists {
		ldflags = append(ldflags, "-l"+filepath.ToSlash(libname))
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

	env = os.Environ()
	env = append(env, cflagsEnv)
	env = append(env, ldflagsEnv)

	fmt.Println(cflagsEnv)
	fmt.Println(ldflagsEnv)
	return args, env, err
}
