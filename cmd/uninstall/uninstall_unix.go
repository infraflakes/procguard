//go:build linux || darwin

package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/blocklist"
	"procguard/internal/config"
)

func platformUninstall() error {
	// Unblock all files
	if err := unblockAll(); err != nil {
		return err
	}

	// Remove systemd service
	if err := removeSystemdService(); err != nil {
		// Don't fail the entire uninstall if this fails, just print a warning.
		fmt.Fprintf(os.Stderr, "Warning: could not remove systemd service: %v\n", err)
	}

	// Remove data files and backup
	return removeDataAndBackup()
}

func unblockAll() error {
	list, err := blocklist.Load()
	if err != nil {
		return fmt.Errorf("could not load blocklist: %w", err)
	}

	for _, name := range list {
		// On Unix, unblocking means restoring execute permissions.
		// This is a simplification; we don't know the original permissions.
		// A more robust implementation would store them.
		if err := os.Chmod(name, 0755); err != nil {
			// Log a warning but continue, as the file may have been moved or deleted.
			fmt.Fprintf(os.Stderr, "Warning: could not unblock %s: %v\n", name, err)
		}
	}

	return nil
}

func removeSystemdService() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if !cfg.SystemdInstalled {
		return nil // Nothing to do
	}

	fmt.Println("Stopping and disabling systemd service...")
	_ = exec.Command("systemctl", "--user", "stop", "procguard.service").Run()
	if err := exec.Command("systemctl", "--user", "disable", "procguard.service").Run(); err != nil {
		return err
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	servicePath := filepath.Join(configDir, "systemd", "user", "procguard.service")

	if err := os.Remove(servicePath); err != nil {
		// Ignore if the file doesn't exist, but return other errors.
		if !os.IsNotExist(err) {
			return err
		}
	}

	return exec.Command("systemctl", "--user", "daemon-reload").Run()
}

func removeDataAndBackup() error {
	fmt.Println("Removing data and backup files...")

	// Remove data from cache
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(cacheDir, "procguard")); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove cache directory: %v\n", err)
	}

	// Remove backup from local share
	dataDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(dataDir, ".local", "share", "procguard")); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove backup directory: %v\n", err)
	}

	return nil
}
