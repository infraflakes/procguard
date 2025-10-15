package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"procguard/internal/app"
	"procguard/internal/data"
	"strings"
	"sync"

	"github.com/bi-zone/go-fileversion"
)

// Server holds the dependencies for the GUI server.
type Server struct {
	Logger          data.Logger
	IsAuthenticated bool
	Mu              sync.Mutex
	db              *sql.DB
}

// NewServer creates a new Server with its dependencies.
func NewServer() (*Server, error) {
	db, err := data.OpenDB()
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Server{
		Logger: data.GetLogger(),
		db:     db,
	}, nil
}

// StartWebServer configures and starts the blocking web server.
func StartWebServer(addr string, registerExtraRoutes func(srv *Server, r *http.ServeMux)) {
	srv, err := NewServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating server:", err)
		os.Exit(1)
	}
	defer func() {
		if err := srv.db.Close(); err != nil {
			srv.Logger.Printf("Failed to close database: %v", err)
		}
	}()

	r := http.NewServeMux()

	srv.registerRoutes(r)
	if registerExtraRoutes != nil {
		registerExtraRoutes(srv, r)
	}

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, srv.authMiddleware(r)); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

// authMiddleware protects all routes except login and assets
func (srv *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.Mu.Lock()
		localIsAuthenticated := srv.IsAuthenticated
		srv.Mu.Unlock()

		if !localIsAuthenticated && r.URL.Path != "/login" && r.URL.Path != "/api/has-password" && r.URL.Path != "/api/login" && r.URL.Path != "/api/set-password" && r.URL.Path != "/api/blocklist" && !strings.HasPrefix(r.URL.Path, "/src/") && !strings.HasPrefix(r.URL.Path, "/dist/") {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (srv *Server) handleAppDetails(w http.ResponseWriter, r *http.Request) {
	exePath := r.URL.Query().Get("path")
	if exePath == "" {
		http.Error(w, "Missing app path", http.StatusBadRequest)
		return
	}

	// Get commercial name
	info, err := fileversion.New(exePath)
	var commercialName string
	if err == nil {
		commercialName = info.FileDescription()
		if commercialName == "" {
			commercialName = info.ProductName()
		}
		if commercialName == "" {
			commercialName = info.OriginalFilename()
		}
	}

	// If we still don't have a good name, use the filename without the extension.
	if commercialName == "" {
		commercialName = strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
	}

	// Get icon
	icon, err := app.GetAppIconAsBase64(exePath)
	if err != nil {
		// Log the error but don't fail the request
		srv.Logger.Printf("Failed to get icon for %s: %v", exePath, err)
	}

	response := struct {
		CommercialName string `json:"commercialName"`
		Icon           string `json:"icon"`
	}{
		CommercialName: commercialName,
		Icon:           icon,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (srv *Server) registerRoutes(r *http.ServeMux) {
	// Handlers
	r.HandleFunc("/logout", srv.handleLogout)

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
	r.HandleFunc("/api/app-details", srv.handleAppDetails)
}
