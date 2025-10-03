//go:build windows

package platform

import "os"

// blockFile renames the file to *.blocked on Windows.
func blockFile(path string) error {
	return os.Rename(path, path+".blocked")
}

// unblockFile renames the file back from *.blocked on Windows.
func unblockFile(path string) error {
	return os.Rename(path+".blocked", path)
}
