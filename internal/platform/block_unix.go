//go:build linux || darwin

package platform

import "os"

// blockFile removes execute permissions on Unix-like systems.
func blockFile(path string) error {
	return os.Chmod(path, 0644)
}

// unblockFile restores execute permissions on Unix-like systems.
func unblockFile(path string) error {
	return os.Chmod(path, 0755)
}
