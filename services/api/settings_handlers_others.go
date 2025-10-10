//go:build !windows

package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetAutostartStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"enabled": false}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleEnableAutostart(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Autostart is not supported on this operating system", http.StatusNotImplemented)
}

func (s *Server) handleDisableAutostart(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Autostart is not supported on this operating system", http.StatusNotImplemented)
}
