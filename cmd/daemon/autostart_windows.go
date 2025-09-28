//go:build windows

package daemon

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/config"
)

const taskName = "ProcGuardDaemon"

// EnsureAutostartTask checks if the autostart task exists and creates it if it doesn't.
// On creation, it copies the executable to a persistent location and points the task there.
func EnsureAutostartTask() {
	// Check if the task already exists.
	err := exec.Command("schtasks", "/query", "/tn", taskName).Run()
	if err == nil {
		return // Task already exists, do nothing.
	}

	fmt.Println("Performing first-time setup for ProcGuard persistence...")

	// 1. Get path of the currently running executable (the source).
	sourcePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
		return
	}

	// 2. Define the hidden backup location (the destination).
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		fmt.Fprintln(os.Stderr, "Could not find LOCALAPPDATA directory.")
		return
	}
	destDir := filepath.Join(localAppData, "ProcGuard")
	destPath := filepath.Join(destDir, "procguard.exe")

	// 3. Copy the executable to the backup location.
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating destination directory:", err)
		return
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening source executable:", err)
		return
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating destination executable:", err)
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error copying executable:", err)
		return
	}

	fmt.Println("Executable backed up to", destPath)

	// 4. Create the scheduled task pointing to the NEW backup location.
	fmt.Println("Creating autostart task...")
	cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", `"`+destPath+`"`, "/sc", "ONLOGON", "/rl", "HIGHEST", "/f")
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating autostart task:", err)
	} else {
		fmt.Println("Successfully created autostart task.")
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to load config:", err)
			return
		}
		cfg.AutostartEnabled = true
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config:", err)
		}
	}
}
