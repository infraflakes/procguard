//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"procguard/internal/config"
)

const taskName = "ProcGuardDaemon"

func checkAutostart() {
	cfg, err := config.Load()
	if err != nil {
		return // Can't load config, so can't check autostart
	}

	if cfg.AutostartEnabled {
		// Check if the task exists
		err := exec.Command("schtasks", "/query", "/tn", taskName).Run()
		if err != nil { // Task likely doesn't exist
			installAutostartTask()
		}
	}
}

func installAutostartTask() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	if cfg.AutostartEnabled {
		fmt.Println("Autostart task already installed.")
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
		os.Exit(1)
	}

	// Create a new task that runs the daemon at logon.
	// The /sc ONLOGON flag runs the task when any user logs on.
	// The /rl HIGHEST flag runs the task with the highest privileges.
	cmdStr := fmt.Sprintf("schtasks /create /tn %s /tr \"%%s\" /sc ONLOGON /rl HIGHEST /f", taskName, exePath)
	if err := exec.Command("cmd", "/C", cmdStr).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating task:", err)
		os.Exit(1)
	}

	cfg.AutostartEnabled = true
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving config:", err)
		os.Exit(1)
	}

	fmt.Println("Autostart task installed.")
}
