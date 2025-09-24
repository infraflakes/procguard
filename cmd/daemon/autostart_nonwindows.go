//go:build !windows

package daemon

// EnsureAutostartTask is a no-op on non-Windows systems.
func EnsureAutostartTask() {
	// This functionality is only for Windows.
}
