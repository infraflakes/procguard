package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"procguard/internal/database"
	"procguard/internal/logger"
	"sync"
)

// Server holds the dependencies for the GUI server.
type Server struct {
	logger          *log.Logger
	isAuthenticated bool
	mu              sync.Mutex
	db              *sql.DB
}

// NewServer creates a new Server with its dependencies.
func NewServer() (*Server, error) {
	db, err := database.OpenDB() // Use OpenDB for read-only clients
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Server{
		logger: logger.Get(),
		db:     db,
	}, nil
}

// StartWebServer configures and starts the blocking web server.
func StartWebServer(addr string) {
	srv, err := NewServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(1)
	}
	defer srv.db.Close()

	r := http.NewServeMux()

	srv.registerRoutes(r)

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, srv.authMiddleware(r)); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

// authMiddleware protects all routes except login and assets
func (srv *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.mu.Lock()
		localIsAuthenticated := srv.isAuthenticated
		srv.mu.Unlock()

		if !localIsAuthenticated && r.URL.Path != "/login" && r.URL.Path != "/api/has-password" && r.URL.Path != "/api/login" && r.URL.Path != "/api/set-password" && r.URL.Path != "/api/blocklist" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (srv *Server) registerRoutes(r *http.ServeMux) {
	// Handlers
	r.HandleFunc("/", srv.handleIndex)
	r.HandleFunc("/login", srv.handleLoginTemplate)
	r.HandleFunc("/logout", srv.handleLogout)
	r.HandleFunc("/ping", srv.handlePing)

	// API routes
	r.HandleFunc("/api/has-password", srv.handleHasPassword)
	r.HandleFunc("/api/login", srv.handleLogin)
	r.HandleFunc("/api/set-password", srv.handleSetPassword)
	r.HandleFunc("/api/search", srv.apiSearch)
	r.HandleFunc("/api/block", srv.apiBlock)
	r.HandleFunc("/api/blocklist", srv.apiBlockList)
	r.HandleFunc("/api/blocklist/clear", srv.apiClearBlocklist)
	r.HandleFunc("/api/blocklist/save", srv.apiSaveBlocklist)
	r.HandleFunc("/api/blocklist/load", srv.apiLoadBlocklist)
	r.HandleFunc("/api/unblock", srv.apiUnblock)
	r.HandleFunc("/api/uninstall", srv.apiUninstall)

	// Web Blocklist API routes
	r.HandleFunc("/api/web-blocklist", srv.handleGetWebBlocklist)
	r.HandleFunc("/api/web-blocklist/add", srv.handleAddWebBlocklist)
	r.HandleFunc("/api/web-blocklist/remove", srv.handleRemoveWebBlocklist)
	r.HandleFunc("/api/web-blocklist/clear", srv.handleClearWebBlocklist)
	r.HandleFunc("/api/web-blocklist/save", srv.handleSaveWebBlocklist)
	r.HandleFunc("/api/web-blocklist/load", srv.handleLoadWebBlocklist)

	// Web Log API routes
	r.HandleFunc("/api/web-logs", srv.handleGetWebLogs)

	// Settings API routes
	r.HandleFunc("/api/settings/autostart/status", srv.handleGetAutostartStatus)
	r.HandleFunc("/api/settings/autostart/enable", srv.handleEnableAutostart)
	r.HandleFunc("/api/settings/autostart/disable", srv.handleDisableAutostart)
}
