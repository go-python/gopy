// Copyright 2015 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	app *commander.Command
)

func init() {
	app = &commander.Command{
		UsageLine: "gopy",
		Subcommands: []*commander.Command{
			gopyMakeCmdGen(),
			gopyMakeCmdBind(),
		},
		Flag: *flag.NewFlagSet("gopy", flag.ExitOnError),
	}
}

func main() {
	err := app.Flag.Parse(os.Args[1:])
	if err != nil {
		log.Printf("error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := app.Flag.Args()
	err = app.Dispatch(args)
	if err != nil {
		log.Printf("error dispatching command: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
