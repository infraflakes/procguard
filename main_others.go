//go:build !windows

package main

import (
	"procguard/cmd"
)

func main() {
	cmd.Execute()
}
