package platform

import (
	"os/exec"
	"path/filepath"
)

// BlockExecutable finds an executable and applies the platform-specific block.
func BlockExecutable(name string) error {
	path, err := findExecutable(name)
	if err != nil {
		return err
	}
	return blockFile(path) // This will be dispatched by build tags
}

// UnblockExecutable finds an executable and reverses the block.
func UnblockExecutable(name string) error {
	path, err := findExecutable(name)
	if err != nil {
		return err
	}
	return unblockFile(path) // This will be dispatched by build tags
}

// findExecutable locates an executable by name.
func findExecutable(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}
	return exec.LookPath(name)
}
