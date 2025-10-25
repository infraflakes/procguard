//go:build windows

package data

import (
	"os/exec"
)

// platformLock sets file permissions on Windows to restrict write access.
// This is a security measure to prevent unauthorized modification of the blocklist files.
// It uses the `icacls` command to:
// 1. Disable inheritance of ACLs from the parent directory (`/inheritance:d`).
// 2. Grant the current user write permissions (`/grant:r`, where `r` means replace).
// 3. Remove the `Everyone` group's permissions, effectively locking the file to the current user.
func platformLock(path string) error {
	return exec.Command("icacls", path, "/inheritance:d", "/grant:r", "%USERNAME%:(W)", "/remove", "Everyone").Run()
}
