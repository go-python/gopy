// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

const (
	defaultPyVersion = "py2"
)

func run(args []string) error {
	app := &commander.Command{
		UsageLine: "gopy",
		Subcommands: []*commander.Command{
			gopyMakeCmdGen(),
			gopyMakeCmdBind(),
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
