//go:build windows

package windows

import (
	"os"
	"os/exec"
	"path/filepath"
)

// blockFile renames to *.blocked (reversible).
func blockFile(path string) error {
	blocked := path + ".blocked"
	return os.Rename(path, blocked)
}

// unblockFile reverses rename.
func unblockFile(path string) error {
	blocked := path + ".blocked"
	return os.Rename(blocked, path)
}

// BlockExecutable finds an executable by name and applies the appropriate
// platform-specific blocking mechanism.
func BlockExecutable(name string) error {
	path, err := findExecutable(name)
	if err != nil {
		return err
	}
	// The actual blocking is handled by a platform-specific function.
	return blockFile(path) // build-tag dispatch
}

// UnblockExecutable finds an executable by name and reverses the blocking mechanism.
func UnblockExecutable(name string) error {
	path, err := findExecutable(name)
	if err != nil {
		return err
	}
	// The actual unblocking is handled by a platform-specific function.
	return unblockFile(path) // build-tag dispatch
}

// findExecutable locates an executable by name, searching in the system's PATH
// and the current working directory. It returns the absolute path to the executable.
func findExecutable(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}
	// exec.LookPath provides a cross-platform way to find executables.
	return exec.LookPath(name)
}
