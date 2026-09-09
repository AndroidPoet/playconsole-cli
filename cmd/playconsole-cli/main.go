package main

import (
	"fmt"
	"os"
	_ "time/tzdata" // vitals needs America/Los_Angeles even on hosts without tzdata

	"github.com/AndroidPoet/playconsole-cli/cmd/playconsole-cli/commands"
)

// Version information (set by build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	commands.SetVersionInfo(version, commit, date)
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
