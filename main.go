package main

import (
	"os"

	"github.com/ErikHellman/coop-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
