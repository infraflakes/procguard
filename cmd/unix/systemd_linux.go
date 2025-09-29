//go:build linux

package unix

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/config"

	"github.com/spf13/cobra"
)

func init() {
	SystemdCmd.AddCommand(systemdInstallCmd, systemdRemoveCmd)
}

var SystemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "Manage systemd user service (Linux only)",
}

var systemdInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and enable the systemd user service",
	RunE:  installSystemdServiceE,
}

var systemdRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Disable and remove the systemd user service",
	RunE:  removeSystemdServiceE,
}

// getServiceFilePath returns the path where the systemd service file should be stored.
func getServiceFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "systemd", "user", "procguard.service"), nil
}

// installSystemdServiceE creates and installs a systemd user service for the ProcGuard daemon.
func installSystemdServiceE(cmd *cobra.Command, args []string) error {
	// First, check if the service is already installed to avoid duplication.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	if cfg.SystemdInstalled {
		fmt.Println("Systemd service already installed.")
		return nil
	}

	// 1. Get path of the currently running executable (the source).
	sourcePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	// 2. Define the hidden backup location in the user's local data directory.
	dataDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not find user home directory: %w", err)
	}
	destDir := filepath.Join(dataDir, ".local", "share", "procguard")
	destPath := filepath.Join(destDir, "procguard")

	// 3. Copy the executable to the backup location.
	fmt.Println("Creating backup of executable...")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error opening source executable: %w", err)
	}
	defer func() {
		if err := sourceFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing source file: %v\n", err)
		}
	}()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating destination executable: %w", err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing destination file: %v\n", err)
		}
	}()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("error copying executable: %w", err)
	}

	// Ensure the backup is executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("error setting executable permission on backup: %w", err)
	}

	fmt.Println("Executable backed up to", destPath)

	// 4. Define the content of the systemd service file, pointing to the backup.
	serviceContent := fmt.Sprintf(`[Unit]
Description=ProcGuard Daemon

[Service]
ExecStart=%s daemon
Restart=always

[Install]
WantedBy=default.target
`, destPath)

	servicePath, err := getServiceFilePath()
	if err != nil {
		return fmt.Errorf("error getting service file path: %w", err)
	}

	// Create the directory for the service file if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
		return fmt.Errorf("error creating systemd directory: %w", err)
	}

	// Write the service file to the systemd user directory.
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("error writing service file: %w", err)
	}

	// Reload the systemd daemon to make it aware of the new service.
	fmt.Println("Reloading systemd user daemon...")
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("error reloading systemd: %w", err)
	}

	// Enable the service to ensure it starts automatically on boot.
	fmt.Println("Enabling procguard service...")
	if err := exec.Command("systemctl", "--user", "enable", "procguard.service").Run(); err != nil {
		return fmt.Errorf("error enabling service: %w", err)
	}

	// Update the configuration to reflect that the service is installed.
	cfg.SystemdInstalled = true
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	fmt.Println("Service installed. Start it with: systemctl --user start procguard.service")
	return nil
}

// removeSystemdServiceE stops, disables, and removes the systemd user service.
func removeSystemdServiceE(cmd *cobra.Command, args []string) error {
	// First, check if the service is actually installed.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	if !cfg.SystemdInstalled {
		fmt.Println("Systemd service not installed.")
		return nil
	}

	// Stop the service if it's running.
	fmt.Println("Stopping procguard service...")
	if err := exec.Command("systemctl", "--user", "stop", "procguard.service").Run(); err != nil {
		// Don't return an error here, as the service might not be running.
		fmt.Fprintln(os.Stderr, "Warning: could not stop service (it may not be running):", err)
	}

	// Disable the service to prevent it from starting on boot.
	fmt.Println("Disabling procguard service...")
	if err := exec.Command("systemctl", "--user", "disable", "procguard.service").Run(); err != nil {
		return fmt.Errorf("error disabling service: %w", err)
	}

	servicePath, err := getServiceFilePath()
	if err != nil {
		return fmt.Errorf("error getting service file path: %w", err)
	}

	// Remove the service file from the systemd user directory.
	if err := os.Remove(servicePath); err != nil {
		return fmt.Errorf("error removing service file: %w", err)
	}

	// Update the configuration to reflect that the service has been removed.
	cfg.SystemdInstalled = false
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	// Reload the systemd daemon to apply the changes.
	fmt.Println("Reloading systemd user daemon...")
	if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("error reloading systemd: %w", err)
	}

	fmt.Println("Service removed.")
	return nil
}
