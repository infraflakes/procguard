//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
)

const taskName = "ProcGuardDaemon"

// EnsureAutostartTask checks if the autostart task exists in the Windows Task Scheduler
// and creates it if it doesn't.
func EnsureAutostartTask() {
	// Check if the task already exists.
	// The `schtasks /query` command returns a non-zero exit code if the task is not found.
	err := exec.Command("schtasks", "/query", "/tn", taskName).Run()

	// If err is nil, the task was found, so we don't need to do anything.
	if err == nil {
		// For quiet operation, you might want to remove this print statement.
		// fmt.Println("Autostart task already exists.")
		return
	}

	// Task not found, so let's create it.
	fmt.Println("Autostart task not found. Creating it now...")
	exePath, err := os.Executable()
	if err != nil {
		// Cannot get executable path, cannot create task.
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
		return
	}

	// Create a new task that runs the executable on user logon.
	// /sc ONLOGON - Runs when any user logs on.
	// /rl HIGHEST - Runs with the highest privileges.
	// /f - Suppresses the confirmation message if the task already exists (useful for updates).
	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", exePath, "/sc", "ONLOGON", "/rl", "HIGHEST", "/f")
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating autostart task:", err)
	} else {
		fmt.Println("Successfully created autostart task.")
	}
}
