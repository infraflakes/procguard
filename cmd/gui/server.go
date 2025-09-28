package gui

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"procguard/internal/logger"
)

// Server holds the dependencies for the GUI server.
type Server struct {
	logger *log.Logger
}

// NewServer creates a new Server with its dependencies.
func NewServer() *Server {
	return &Server{
		logger: logger.Get(),
	}
}

// StartWebServer configures and starts the blocking web server.
func StartWebServer(addr string) {
	s := NewServer()
	r := http.NewServeMux()

	// Middleware to protect all routes except login and assets
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			localIsAuthenticated := isAuthenticated
			mu.Unlock()

			if !localIsAuthenticated && r.URL.Path != "/login" && r.URL.Path != "/api/has-password" && r.URL.Path != "/api/login" && r.URL.Path != "/api/set-password" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	s.registerRoutes(r)

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, authMiddleware(r)); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

func (s *Server) registerRoutes(r *http.ServeMux) {
	// Handlers
	r.HandleFunc("/", s.handleIndex)
	r.HandleFunc("/login", s.handleLoginTemplate)
	r.HandleFunc("/logout", s.handleLogout)
	r.HandleFunc("/ping", s.handlePing)

	// API routes
	r.HandleFunc("/api/has-password", s.handleHasPassword)
	r.HandleFunc("/api/login", s.handleLogin)
	r.HandleFunc("/api/set-password", s.handleSetPassword)
	r.HandleFunc("/api/search", s.apiSearch)
	r.HandleFunc("/api/block", s.apiBlock)
	r.HandleFunc("/api/blocklist", s.apiBlockList)
	r.HandleFunc("/api/blocklist/clear", s.apiClearBlocklist)
	r.HandleFunc("/api/blocklist/save", s.apiSaveBlocklist)
	r.HandleFunc("/api/blocklist/load", s.apiLoadBlocklist)
	r.HandleFunc("/api/unblock", s.apiUnblock)
}