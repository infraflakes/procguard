//go:build linux

package unix

import (
	"fmt"
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
	Run:   installSystemdService,
}

var systemdRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Disable and remove the systemd user service",
	Run:   removeSystemdService,
}

// getServiceFilePath returns the path where the systemd service file should be stored.
func getServiceFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "systemd", "user", "procguard.service"), nil
}

// installSystemdService creates and installs a systemd user service for the ProcGuard daemon.
func installSystemdService(cmd *cobra.Command, args []string) {
	// First, check if the service is already installed to avoid duplication.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	if cfg.SystemdInstalled {
		fmt.Println("Systemd service already installed.")
		return
	}

	// Get the path to the currently running executable to use in the service file.
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
		os.Exit(1)
	}

	// Define the content of the systemd service file.
	serviceContent := fmt.Sprintf(`[Unit]
Description=ProcGuard Daemon

[Service]
ExecStart=%s daemon
Restart=always

[Install]
WantedBy=default.target
`, exePath)

	servicePath, err := getServiceFilePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting service file path:", err)
		os.Exit(1)
	}

	// Create the directory for the service file if it doesn't exist.
	if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating systemd directory:", err)
		os.Exit(1)
	}

	// Write the service file to the systemd user directory.
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing service file:", err)
		os.Exit(1)
	}

	// Reload the systemd daemon to make it aware of the new service.
	fmt.Println("Reloading systemd user daemon...")
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	// Enable the service to ensure it starts automatically on boot.
	fmt.Println("Enabling procguard service...")
	if err := exec.Command("systemctl", "--user", "enable", "procguard.service").Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error enabling service:", err)
		os.Exit(1)
	}

	// Update the configuration to reflect that the service is installed.
	cfg.SystemdInstalled = true
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving config:", err)
		os.Exit(1)
	}

	fmt.Println("Service installed. Start it with: systemctl --user start procguard.service")
}

// removeSystemdService stops, disables, and removes the systemd user service.
func removeSystemdService(cmd *cobra.Command, args []string) {
	// First, check if the service is actually installed.
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	if !cfg.SystemdInstalled {
		fmt.Println("Systemd service not installed.")
		return
	}

	// Stop the service if it's running.
	fmt.Println("Stopping procguard service...")
	exec.Command("systemctl", "--user", "stop", "procguard.service").Run()

	// Disable the service to prevent it from starting on boot.
	fmt.Println("Disabling procguard service...")
	exec.Command("systemctl", "--user", "disable", "procguard.service").Run()

	servicePath, err := getServiceFilePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting service file path:", err)
		os.Exit(1)
	}

	// Remove the service file from the systemd user directory.
	if err := os.Remove(servicePath); err != nil {
		fmt.Fprintln(os.Stderr, "Error removing service file:", err)
	}

	// Update the configuration to reflect that the service has been removed.
	cfg.SystemdInstalled = false
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving config:", err)
		os.Exit(1)
	}

	// Reload the systemd daemon to apply the changes.
	fmt.Println("Reloading systemd user daemon...")
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Println("Service removed.")
}
