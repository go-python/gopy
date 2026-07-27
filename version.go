// Version is bumped by release-please via the x-release-please-version
// marker below; do not hand-edit the value on that line.
//
// GitCommit and VersionDate are stamped at build time via `go build -ldflags`
// (see Makefile and .goreleaser.yaml), so they default to "unknown" here.

package main

var (
	Version     = "v0.4.10" // x-release-please-version
	GitCommit   = "unknown"
	VersionDate = "unknown"
)
