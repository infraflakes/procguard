package api

import (
	"net/http"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.mu.Lock()
	localIsAuthenticated := s.isAuthenticated
	s.mu.Unlock()

	if !localIsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(dashboardHTML); err != nil {
		s.logger.Printf("Error writing response: %v", err)
	}
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
