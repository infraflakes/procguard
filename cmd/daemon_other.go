//go:build !windows

package cmd

func checkAutostart() {
	// This is a no-op on non-Windows systems.
}
