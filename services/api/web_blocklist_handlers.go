package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/blocklist/webblocklist"
)

func (s *Server) handleGetWebBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := webblocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load web blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleAddWebBlocklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := webblocklist.Add(req.Domain); err != nil {
		http.Error(w, "Failed to add to web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleRemoveWebBlocklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := webblocklist.Remove(req.Domain); err != nil {
		http.Error(w, "Failed to remove from web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}
