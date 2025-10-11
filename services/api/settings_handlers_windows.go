//go:build windows

package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/config"
	"procguard/internal/daemon"
)

func (s *Server) handleGetAutostartStatus(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"enabled": cfg.AutostartEnabled}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleEnableAutostart(w http.ResponseWriter, r *http.Request) {
	_, err := daemon.EnsureAutostart()
	if err != nil {
		http.Error(w, "Failed to enable autostart: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDisableAutostart(w http.ResponseWriter, r *http.Request) {
	if err := daemon.RemoveAutostart(); err != nil {
		http.Error(w, "Failed to disable autostart: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
