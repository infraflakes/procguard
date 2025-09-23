//go:build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(systemdCmd)
	systemdCmd.AddCommand(systemdInstallCmd, systemdRemoveCmd)
}

var systemdCmd = &cobra.Command{
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

func getServiceFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "systemd", "user", "procguard.service"), nil
}

func installSystemdService(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	if cfg.SystemdInstalled {
		fmt.Println("Systemd service already installed.")
		return
	}
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
		os.Exit(1)
	}

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

	if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Error creating systemd directory:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing service file:", err)
		os.Exit(1)
	}

	fmt.Println("Reloading systemd user daemon...")
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Println("Enabling procguard service...")
	if err := exec.Command("systemctl", "--user", "enable", "procguard.service").Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error enabling service:", err)
		os.Exit(1)
	}

	cfg.SystemdInstalled = true
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving config:", err)
		os.Exit(1)
	}

	fmt.Println("Service installed. Start it with: systemctl --user start procguard.service")
}

func removeSystemdService(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	if !cfg.SystemdInstalled {
		fmt.Println("Systemd service not installed.")
		return
	}

	fmt.Println("Stopping procguard service...")
	exec.Command("systemctl", "--user", "stop", "procguard.service").Run()

	fmt.Println("Disabling procguard service...")
	exec.Command("systemctl", "--user", "disable", "procguard.service").Run()

	servicePath, err := getServiceFilePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting service file path:", err)
		os.Exit(1)
	}

	if err := os.Remove(servicePath); err != nil {
		fmt.Fprintln(os.Stderr, "Error removing service file:", err)
	}

	cfg.SystemdInstalled = false
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving config:", err)
		os.Exit(1)
	}

	fmt.Println("Reloading systemd user daemon...")
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Println("Service removed.")
}
