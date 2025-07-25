package main

import (
	"github.com/voidkey-labs/cli/cmd"
)

// These variables are set by goreleaser during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}