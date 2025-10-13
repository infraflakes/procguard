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
	if err := Templates.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
		logger.Printf("Error executing dashboard template: %v", err)
	}
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HandleLoginTemplate(logger data.Logger, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		logger.Printf("Error executing login template: %v", err)
	}
}
