//go:build !windows

package daemon

import (
	"os"
	"procguard/internal/ignore"

	"github.com/shirou/gopsutil/v3/process"
)

// shouldLogProcess determines if a process should be logged based on platform-specific rules.
func shouldLogProcess(p *process.Process) bool {
	name, err := p.Name()
	if err != nil || name == "" {
		return false // Skip processes with no name
	}

	// Universal check: ignore self
	if p.Pid == int32(os.Getpid()) {
		return false
	}

	parent, err := p.Parent()
	if err != nil {
		return false // Skip processes with no parent
	}
	parentName, _ := parent.Name()

	// Do not log a process if its parent has the same name.
	if name == parentName {
		return false
	}

	// Linux-specific checks
	uids, err := p.Uids()
	if err != nil || len(uids) == 0 || uids[0] < 1000 {
		return false // Skip system users
	}
	if ignore.IsIgnored(name, ignore.DefaultLinux) || ignore.IsIgnored(parentName, ignore.DefaultLinux) {
		return false
	}

	return true
}
