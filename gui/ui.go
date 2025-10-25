package gui

import (
	"net/http"
	"procguard/internal/data"
	"sync"
)

// HandleIndex serves the main index page of the web UI.
// It checks for authentication and redirects to the login page if the user is not authenticated.
// TODO: This handler has a large number of parameters. It could be refactored to be a method on a struct that holds the dependencies.
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
	if err := Templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		logger.Printf("Error executing dashboard template: %v", err)
	}
}

// HandlePing is a simple health check endpoint that returns a 200 OK status.
func HandlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// HandleLoginTemplate serves the login page.
func HandleLoginTemplate(logger data.Logger, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		logger.Printf("Error executing login template: %v", err)
	}
}
