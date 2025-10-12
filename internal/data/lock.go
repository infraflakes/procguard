//go:build windows

package data

import (
	"os/exec"
)

func platformLock(path string) error {
	// **inherit parent ACL** + **remove Everyone write** but **keep owner**
	return exec.Command("icacls", path, "/inheritance:d", "/grant:r", "%USERNAME%:(W)", "/remove", "Everyone").Run()
}
