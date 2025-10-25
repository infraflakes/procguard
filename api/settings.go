//go:build windows

package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/daemon"
	"procguard/internal/data"
)

// handleGetAutostartStatus returns the current status of the autostart setting.
func (s *Server) handleGetAutostartStatus(w http.ResponseWriter, r *http.Request) {
	cfg, err := data.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"enabled": cfg.AutostartEnabled}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleEnableAutostart enables the autostart feature for the application.
func (s *Server) handleEnableAutostart(w http.ResponseWriter, r *http.Request) {
	_, err := daemon.EnsureAutostart()
	if err != nil {
		http.Error(w, "Failed to enable autostart: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handleDisableAutostart disables the autostart feature for the application.
func (s *Server) handleDisableAutostart(w http.ResponseWriter, r *http.Request) {
	if err := daemon.RemoveAutostart(); err != nil {
		http.Error(w, "Failed to disable autostart: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
