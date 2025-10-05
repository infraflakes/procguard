//go:build windows

package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/blocklist"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const taskName = "ProcGuardDaemon"
const hostName = "com.nixuris.procguard"

func platformUninstall() error {
	// Unblock all files
	if err := unblockAll(); err != nil {
		return err
	}

	// Remove Task Scheduler task
	if err := removeTask(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove task scheduler entry: %v\n", err)
	}

	// Remove Native Messaging Host configuration
	if err := removeNativeHost(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove native messaging host configuration: %v\n", err)
	}

	// Remove data files
	if err := removeDataFiles(); err != nil {
		return err
	}

	// Remove backup executable
	return removeBackup()
}

func removeNativeHost() error {
	fmt.Println("Removing Native Messaging Host configuration...")

	// Delete the registry key.
	keyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + hostName
	if err := registry.DeleteKey(registry.CURRENT_USER, keyPath); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Delete the manifest file.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(cacheDir, "procguard", "procguard.json")
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func unblockAll() error {
	list, err := blocklist.Load()
	if err != nil {
		return fmt.Errorf("could not load blocklist: %w", err)
	}

	for _, name := range list {
		// On Windows, unblocking means renaming the file.
		if strings.HasSuffix(name, ".blocked") {
			newName := strings.TrimSuffix(name, ".blocked")
			if err := os.Rename(name, newName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not unblock %s: %v\n", name, err)
			}
		}
	}

	return nil
}

func removeTask() error {
	fmt.Println("Removing Task Scheduler task...")
	// The /f flag is to force the deletion.
	return exec.Command("schtasks", "/delete", "/tn", taskName, "/f").Run()
}

func removeDataFiles() error {
	fmt.Println("Removing data files...")
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	procguardDir := filepath.Join(cacheDir, "procguard")
	logsDir := filepath.Join(procguardDir, "logs")

	if err := os.RemoveAll(logsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not remove logs directory: %v\n", err)
	}
	return os.RemoveAll(procguardDir)
}

func removeBackup() error {
	fmt.Println("Removing backup executable...")
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	return os.RemoveAll(filepath.Join(localAppData, "ProcGuard"))
}
