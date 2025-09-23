//go:build linux || darwin
package block

import (
	"os"
)

// blockFile removes **execute** bit (keeps read/write).
func blockFile(path string) error {
	return os.Chmod(path, 0644)
}

// unblockFile restores user execute.
func unblockFile(path string) error {
	return os.Chmod(path, 0755)
}

// platformLock is a no-op on Unix (0600 already set).
func platformLock(path string) error { return nil }
