package main

import (
	"os"

	"github.com/ErikHellman/coop-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
