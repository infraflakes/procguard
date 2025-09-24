package gui

import (
	"os"
	"os/exec"
)

// runProcGuardCommand is a helper to safely create an *exec.Cmd for the procguard binary.
// It finds the absolute path to the current executable to ensure commands work
// regardless of the system's PATH.
func runProcGuardCommand(args ...string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return exec.Command(exe, args...), nil
}
