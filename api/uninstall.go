package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"procguard/internal/auth"
	"procguard/internal/data"
	"procguard/internal/daemon"
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
		s.db.Close()
		time.Sleep(1 * time.Second)
		killOtherProcGuardProcesses()
		time.Sleep(2 * time.Second) // Give processes time to die
		unblockAll()
		daemon.RemoveAutostart()
		removeNativeHost()

		// Remove all application data, logs, and backups from LOCALAPPDATA
		fmt.Println("Removing all application data...")
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			appDataDir := filepath.Join(localAppData, appName)
			os.RemoveAll(appDataDir)
		}

		// Remove cache directory
		cacheDir, err := os.UserCacheDir()
		if err == nil {
			procguardCacheDir := filepath.Join(cacheDir, "procguard")
			os.RemoveAll(procguardCacheDir)
		}

		os.Exit(0)
	}()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func killOtherProcGuardProcesses() {
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
			p.Kill()
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
			os.Rename(name, newName)
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
