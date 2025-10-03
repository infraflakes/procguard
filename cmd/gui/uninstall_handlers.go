package gui

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"procguard/internal/auth"
	"procguard/internal/config"
	"time"
)

func (s *Server) apiUninstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPasswordHash(req.Password, cfg.PasswordHash) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		http.Error(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}

	// We don't run this in a goroutine because we want the server to become
	// unresponsive as it's being uninstalled.
	cmd := exec.Command(exePath, "uninstall", "--force-no-prompt")
	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start uninstall process", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}

	// Exit the application to allow the uninstall to complete.
	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}
