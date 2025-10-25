package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/auth"
	"procguard/internal/data"
)

// handleLogout handles the user logout.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.Mu.Lock()
	s.IsAuthenticated = false
	s.Mu.Unlock()
	http.Redirect(w, r, "/login", http.StatusFound)
}

// handleHasPassword checks if a password has been set for the application.
// It returns a JSON response with a boolean `hasPassword` field.
func (s *Server) handleHasPassword(w http.ResponseWriter, r *http.Request) {
	cfg, err := data.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}
	hasPassword := cfg.PasswordHash != ""
	if err := json.NewEncoder(w).Encode(map[string]bool{"hasPassword": hasPassword}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleLogin handles the user login.
// It expects a JSON request with a `password` field.
// It returns a JSON response with a boolean `success` field.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
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

	if auth.CheckPasswordHash(req.Password, cfg.PasswordHash) {
		s.Mu.Lock()
		s.IsAuthenticated = true
		s.Mu.Unlock()
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
			s.Logger.Printf("Error encoding response: %v", err)
		}
	} else {
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": false}); err != nil {
			s.Logger.Printf("Error encoding response: %v", err)
		}
	}
}

// handleSetPassword handles the initial password setup.
// It expects a JSON request with a `password` field.
// It returns a JSON response with a boolean `success` field.
func (s *Server) handleSetPassword(w http.ResponseWriter, r *http.Request) {
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

	if cfg.PasswordHash != "" {
		http.Error(w, "Password already set", http.StatusForbidden)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	cfg.PasswordHash = hash
	if err := cfg.Save(); err != nil {
		http.Error(w, "Failed to save password", http.StatusInternalServerError)
		return
	}

	s.Mu.Lock()
	s.IsAuthenticated = true
	s.Mu.Unlock()
	if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}
