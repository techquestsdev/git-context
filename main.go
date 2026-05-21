package main

import (
	"fmt"
	"os"

	"github.com/techquestsdev/git-context/cmd"
)

// version is set at build time via goreleaser's
// `-X main.version={{.Version}}` ldflag. Defaults to "dev" for local
// `go build` / `go install` invocations.
var version = "dev"

func main() {
	cmd.SetVersion(version)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
