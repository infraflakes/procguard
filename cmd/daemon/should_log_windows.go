//go:build windows

package daemon

import (
	"os"
	"procguard/internal/ignore"
	"procguard/internal/winutil"

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

	// Windows-specific checks
	il, err := winutil.GetProcessIntegrityLevel(uint32(p.Pid))
	if err == nil && il >= winutil.SECURITY_MANDATORY_SYSTEM_RID {
		return false // Skip system/high integrity processes
	}
	if ignore.IsIgnored(name, ignore.DefaultWindows) || ignore.IsIgnored(parentName, ignore.DefaultWindows) {
		return false
	}

	return true
}
