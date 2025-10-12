package gui

import (
	"net/http"
	"procguard/internal/data"
	"sync"
)

func HandleIndex(mu *sync.Mutex, isAuthenticated bool, logger data.Logger, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	mu.Lock()
	localIsAuthenticated := isAuthenticated
	mu.Unlock()

	if !localIsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(DashboardHTML); err != nil {
		logger.Printf("Error writing response: %v", err)
	}
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HandleLoginTemplate(logger data.Logger, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(LoginHTML); err != nil {
		logger.Printf("Error writing response: %v", err)
	}
}

