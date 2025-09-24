//go:build linux || darwin

package cmd

import (
	"procguard/cmd/unix"
)

func init() {
	rootCmd.AddCommand(unix.SystemdCmd)
}
