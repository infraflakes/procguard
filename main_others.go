//go:build !windows

package main

import (
	"os"
	"strings"
	"procguard/cmd"
	"procguard/internal/native"
)

func main() {
	// When launched by Chrome, the first argument is the extension's origin.
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		native.Run()
		return
	}

	cmd.Execute()
}
