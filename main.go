package main

//go:generate go run github.com/akavel/rsrc -manifest procguard.manifest -o rsrc.syso

import (
	"os"
	"procguard/cmd"
)

func main() {
	// If no command-line arguments are provided, this is a default run (e.g., double-click).
	if len(os.Args) == 1 {
		cmd.HandleDefaultStartup()
	} else {
		// Otherwise, execute the CLI command provided by the user.
		cmd.Execute()
	}
}
