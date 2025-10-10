//go:build windows

package daemon

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"procguard/internal/config"

	"golang.org/x/sys/windows/registry"
)

const appName = "ProcGuard"

// EnsureAutostart creates a registry entry and returns the path to the persistent executable.
func EnsureAutostart() (string, error) {
	// The path to the executable in the persistent location
	destPath, err := copyExecutableToAppData()
	if err != nil {
		return "", fmt.Errorf("failed to set up persistent executable: %w", err)
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return destPath, fmt.Errorf("failed to open Run registry key: %w", err)
	}
	defer key.Close()

	// Check if the value already exists and is correct.
	currentPath, _, err := key.GetStringValue(appName)
	if err == nil && currentPath == destPath {
		return destPath, nil // Entry already exists and is correct.
	}

	fmt.Println("Performing first-time setup for ProcGuard persistence...")

	// Set the registry value to point to the persistent executable path.
	if err := key.SetStringValue(appName, destPath); err != nil {
		return destPath, fmt.Errorf("failed to set startup registry key: %w", err)
	}

	// Update the config file to reflect the change
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config to update autostart status:", err)
	} else {
		cfg.AutostartEnabled = true
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config to update autostart status:", err)
		}
	}

	fmt.Println("Successfully created startup registry entry.")
	return destPath, nil
}

// RemoveAutostart removes the registry entry that starts the application on logon.
func RemoveAutostart() error {
	fmt.Println("Removing autostart registry entry...")
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return nil // Key doesn't exist, nothing to do.
		}
		return err
	}
	defer key.Close()

	// Delete the value. If it doesn't exist, this will return an error that we can ignore.
	if err := key.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Update the config file to reflect the change
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config to update autostart status:", err)
	} else {
		cfg.AutostartEnabled = false
		if err := cfg.Save(); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to save config to update autostart status:", err)
		}
	}

	return nil
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