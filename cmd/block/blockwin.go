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

func platformLock(path string) error {
	// **inherit parent ACL** + **remove Everyone write** but **keep owner**
	return exec.Command("icacls", path, "/inheritance:d", "/grant:r", "%USERNAME%:(W)", "/remove", "Everyone").Run()
}
