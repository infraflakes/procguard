//go:build !windows

package cmd

// HandleDefaultStartup is a no-op on non-Windows systems.
func HandleDefaultStartup() {
	// This functionality is only for Windows.
}
