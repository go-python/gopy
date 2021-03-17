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

	"github.com/go-python/gopy/bind"
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
func NewBuildCfg() *BuildCfg {
	var cfg BuildCfg
	cfg.Cmd = argStr()
	return &cfg
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
