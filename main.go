// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/pkg/errors"

	"github.com/rudderlabs/gopy/bind"
)

// BuildCfg contains command options and binding generation options
type BuildCfg struct {
	bind.BindCfg

	// include symbols in output
	Symbols bool

	// suppress warning messages, which may be expected
	NoWarn bool

	// do not generate a Makefile, e.g., when called from Makefile
	NoMake bool
}

// NewBuildCfg returns a newly constructed build config
func NewBuildCfg(flagSet *flag.FlagSet) *BuildCfg {
	var cfg BuildCfg
	cfg.Cmd = argStr()

	cfg.OutputDir = flagSet.Lookup("output").Value.Get().(string)
	cfg.Name = flagSet.Lookup("name").Value.Get().(string)
	cfg.Main = flagSet.Lookup("main").Value.Get().(string)
	cfg.VM = flagSet.Lookup("vm").Value.Get().(string)
	cfg.RenameCase = flagSet.Lookup("rename").Value.Get().(bool)
	cfg.NoWarn = flagSet.Lookup("no-warn").Value.Get().(bool)
	cfg.NoMake = flagSet.Lookup("no-make").Value.Get().(bool)
	cfg.PkgPrefix = flagSet.Lookup("package-prefix").Value.Get().(string)
	cfg.NoPyExceptions = flagSet.Lookup("no-exceptions").Value.Get().(bool)
	cfg.ModPathGoErr2PyEx = flagSet.Lookup("gomod-2pyex").Value.Get().(string)
	cfg.UsePyTuple4VE = flagSet.Lookup("use-pytuple-4ve").Value.Get().(bool)

	if cfg.ModPathGoErr2PyEx == "" {
		cfg.ModPathGoErr2PyEx = "github.com/rudderlabs/gopy/goerr2pyex/"
	}

	if cfg.VM == "" {
		cfg.VM = "python"
	}

	return &cfg
}

func AddCommonCmdFlags(flagSet *flag.FlagSet) {
	flagSet.String("vm", "python", "path to python interpreter")
	flagSet.String("output", "", "output directory for root of package")
	flagSet.String("name", "", "name of output package (otherwise name of first package is used)")
	flagSet.String("main", "", "code string to run in the go main()/GoPyInit() function(s) in the cgo library "+
		"-- defaults to GoPyMainRun() but typically should be overriden")

	flagSet.Bool("rename", false, "rename Go symbols to python PEP snake_case")
	flagSet.Bool("no-warn", false, "suppress warning messages, which may be expected")
	flagSet.Bool("no-make", false, "do not generate a Makefile, e.g., when called from Makefile")

	flagSet.Bool("no-exceptions", false, "Don't throw python exceptions when the last return is in error.")
	flagSet.Bool("use-pytuple-4ve", false, "When Go returns (value, err) with err=nil, return (value, ) tuple to python.")
	flagSet.String("gomod-2pyex", "", "Go module to use to translate Go errors to Python exceptions.")
}

func run(args []string) error {
	app := &commander.Command{
		UsageLine: "gopy",
		Subcommands: []*commander.Command{
			gopyMakeCmdGen(),
			gopyMakeCmdBuild(),
			gopyMakeCmdPkg(),
			gopyMakeCmdExe(),
		},
		Flag: *flag.NewFlagSet("gopy", flag.ExitOnError),
	}

	err := app.Flag.Parse(args)
	if err != nil {
		return fmt.Errorf("could not parse flags: %v", err)
	}

	appArgs := app.Flag.Args()
	err = app.Dispatch(appArgs)
	if err != nil {
		return fmt.Errorf("error dispatching command: %v", err)
	}
	return nil
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func copyCmd(src, dst string) error {
	srcf, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "could not open source for copy")
	}
	defer srcf.Close()

	os.MkdirAll(path.Dir(dst), 0755)

	dstf, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, "could not create destination for copy")
	}
	defer dstf.Close()

	_, err = io.Copy(dstf, srcf)
	if err != nil {
		return errors.Wrap(err, "could not copy bytes to destination")
	}

	err = dstf.Sync()
	if err != nil {
		return errors.Wrap(err, "could not synchronize destination")
	}

	return dstf.Close()
}
