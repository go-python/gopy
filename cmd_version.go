// Copyright 2026 The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func gopyMakeCmdVersion() *commander.Command {
	return &commander.Command{
		Run:       gopyRunCmdVersion,
		UsageLine: "version",
		Short:     "print gopy version information",
		Long: `
version prints the gopy version, git commit, and build date embedded in this binary at release time.

ex:
 $ gopy version
`,
		Flag: *flag.NewFlagSet("gopy-version", flag.ExitOnError),
	}
}

func gopyRunCmdVersion(cmdr *commander.Command, args []string) error {
	fmt.Printf("gopy version %s (commit %s, built %s)\n", Version, GitCommit, VersionDate)
	return nil
}
