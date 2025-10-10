//go:build windows

package daemon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const appName = "ProcGuard"

// EnsureAutostart creates a registry entry in the current user's Run key to start the application on logon.
func EnsureAutostart() {
	// The path to the executable in the persistent location
	destPath, err := copyExecutableToAppData()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set up persistent executable:", err)
		return
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open Run registry key:", err)
		return
	}
	defer key.Close()

	// Check if the value already exists and is correct.
	currentPath, _, err := key.GetStringValue(appName)
	if err == nil && currentPath == destPath {
		return // Entry already exists and is correct.
	}

	fmt.Println("Performing first-time setup for ProcGuard persistence...")

	// Set the registry value to point to the persistent executable path.
	if err := key.SetStringValue(appName, destPath); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to set startup registry key:", err)
		return
	}

	fmt.Println("Successfully created startup registry entry.")
}

// copyExecutableToAppData copies the current executable to a hidden, persistent location in LOCALAPPDATA.
// It returns the path to the new executable.
func copyExecutableToAppData() (string, error) {
	sourcePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %w", err)
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	destDir := filepath.Join(localAppData, appName)
	destPath := filepath.Join(destDir, "ProcGuardSvc.exe")

	// If the file already exists, no need to copy again.
	if _, err := os.Stat(destPath); err == nil {
		return destPath, nil
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("error creating destination directory: %w", err)
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("error opening source executable: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("error creating destination executable: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("error copying executable: %w", err)
	}

	fmt.Println("Executable backed up to", destPath)
	return destPath, nil
}
