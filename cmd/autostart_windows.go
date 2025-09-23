//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"procguard/internal/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(autostartCmd)
	autostartCmd.AddCommand(autostartInstallCmd, autostartRemoveCmd)
}

var autostartCmd = &cobra.Command{
	Use:   "windows-autostart",
	Short: "Manage Windows autostart task (Windows only)",
}

var autostartInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the autostart task",
	Run:   installAutostartTask,
}

var autostartRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove the autostart task",
	Run:   removeAutostartTask,
}

func installAutostartTask(cmd *cobra.Command, args []string) {
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

func removeAutostartTask(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	if !cfg.AutostartEnabled {
		fmt.Println("Autostart task not installed.")
		return
	}

	// Delete the task.
	cmdStr := fmt.Sprintf("schtasks /delete /tn %s /f", taskName)
	if err := exec.Command("cmd", "/C", cmdStr).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error deleting task:", err)
		os.Exit(1)
	}

	cfg.AutostartEnabled = false
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving config:", err)
		os.Exit(1)
	}

	fmt.Println("Autostart task removed.")
}
