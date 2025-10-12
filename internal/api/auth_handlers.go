package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/config"
	"procguard/gui"
)

func (s *Server) handleLoginTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(gui.LoginHTML); err != nil {
		s.Logger.Printf("Error writing response: %v", err)
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.Mu.Lock()
	s.IsAuthenticated = false
	s.Mu.Unlock()
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handleHasPassword(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}
	hasPassword := cfg.PasswordHash != ""
	if err := json.NewEncoder(w).Encode(map[string]bool{"hasPassword": hasPassword}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
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

	if CheckPasswordHash(req.Password, cfg.PasswordHash) {
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

func (s *Server) handleSetPassword(w http.ResponseWriter, r *http.Request) {
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

	if cfg.PasswordHash != "" {
		http.Error(w, "Password already set", http.StatusForbidden)
		return
	}

	hash, err := HashPassword(req.Password)
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
