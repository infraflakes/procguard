package api

import (
	"encoding/json"
	"fmt"

	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/auth"
	"procguard/internal/daemon"
	"procguard/internal/data"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows/registry"
)

const appName = "ProcGuard"
const hostName = "com.nixuris.procguard"

func (s *Server) apiUninstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg, err := data.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPasswordHash(req.Password, cfg.PasswordHash) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	go func() {
		// Close the logger to release file handles
		s.Logger.Close()
		// Close the database connection
		if err := s.db.Close(); err != nil {
			// We can't use the logger here, so just print to stderr
			fmt.Fprintf(os.Stderr, "Failed to close database: %v\n", err)
		}

		// Kill other ProcGuard processes
		killOtherProcGuardProcesses(s.Logger)

		if err := unblockAll(); err != nil {
			// Log to stderr since the logger is closed
			fmt.Fprintf(os.Stderr, "Failed to unblock all files: %v\n", err)
		}

		// Perform other cleanup tasks that don't involve file deletion
		if err := daemon.RemoveAutostart(); err != nil {
			// Log to stderr since the logger is closed
			fmt.Fprintf(os.Stderr, "Failed to remove autostart: %v\n", err)
		}
		if err := removeNativeHost(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove native host: %v\n", err)
		}

		// Create and launch the self-deleting batch script
		if err := selfDelete(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initiate self-deletion: %v\n", err)
		}

		// Exit the application
		os.Exit(0)
	}()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		// The logger is closed, so just print to stderr
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
	}
}

func selfDelete() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("self-deletion is currently implemented only for Windows")
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return fmt.Errorf("could not find LOCALAPPDATA directory")
	}
	appDataDir := filepath.Join(localAppData, appName)

	// Create a temporary batch file
	tempDir := os.TempDir()
	batchFileName := fmt.Sprintf("delete_procguard_%d.bat", time.Now().UnixNano())
	batchFilePath := filepath.Join(tempDir, batchFileName)

	// The batch script content:
	// 1. Wait for the main process to exit.
	// 2. Delete the application data directory.
	// 3. Delete the batch file itself.
	batchContent := fmt.Sprintf(`
@echo off
timeout /t 2 /nobreak > nul
rmdir /s /q "%s"
del "%s"
`, appDataDir, batchFilePath)

	err := os.WriteFile(batchFilePath, []byte(batchContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write batch file: %w", err)
	}

	// Execute the batch file in a detached process
	cmd := exec.Command("cmd.exe", "/C", batchFilePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start batch process: %w", err)
	}

	return nil
}

func killOtherProcGuardProcesses(logger data.Logger) {
	currentPid := os.Getpid()
	procs, err := process.Processes()
	if err != nil {
		return
	}

	for _, p := range procs {
		if p.Pid == int32(currentPid) {
			continue
		}

		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.HasPrefix(strings.ToLower(name), "procguard") {
			if err := p.Kill(); err != nil {
				logger.Printf("Failed to kill process %s: %v", name, err)
			}
		}
	}
}

func unblockAll() error {
	list, err := data.LoadApp()
	if err != nil {
		return fmt.Errorf("could not load blocklist: %w", err)
	}

	for _, name := range list {
		if strings.HasSuffix(name, ".blocked") {
			newName := strings.TrimSuffix(name, ".blocked")
			if err := os.Rename(name, newName); err != nil {
				// Log the error but continue trying to unblock other files
				data.GetLogger().Printf("Failed to unblock file %s: %v", name, err)
			}
		}
	}

	return nil
}

func removeNativeHost() error {
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
