package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"procguard/internal/auth"
	"procguard/internal/daemon"
	"procguard/internal/data"
	"strings"
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
		if err := s.db.Close(); err != nil {
			s.Logger.Printf("Failed to close database: %v", err)
		}
		time.Sleep(1 * time.Second)
		killOtherProcGuardProcesses(s.Logger)
		time.Sleep(2 * time.Second) // Give processes time to die
		if err := unblockAll(); err != nil {
			s.Logger.Printf("Failed to unblock all files: %v", err)
		}
		if err := daemon.RemoveAutostart(); err != nil {
			s.Logger.Printf("Failed to remove autostart: %v", err)
		}
		if err := removeNativeHost(); err != nil {
			s.Logger.Printf("Failed to remove native host: %v", err)
		}

		// Remove all application data, logs, and backups from LOCALAPPDATA
		fmt.Println("Removing all application data...")
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			appDataDir := filepath.Join(localAppData, appName)
			if err := os.RemoveAll(appDataDir); err != nil {
				s.Logger.Printf("Failed to remove app data directory: %v", err)
			}
		}

		// Remove cache directory
		cacheDir, err := os.UserCacheDir()
		if err == nil {
			procguardCacheDir := filepath.Join(cacheDir, "procguard")
			if err := os.RemoveAll(procguardCacheDir); err != nil {
				s.Logger.Printf("Failed to remove cache directory: %v", err)
			}
		}

		os.Exit(0)
	}()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
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
