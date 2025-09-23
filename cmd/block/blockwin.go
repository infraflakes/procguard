//go:build windows
package block

import (
	"os"
	"os/exec"
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

// platformLock denies **write** to Everyone except owner.
func platformLock(path string) error {
	return exec.Command("icacls", path, "/deny", "Everyone:(W)").Run()
}
