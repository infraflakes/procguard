package gui

import (
	"fmt"
	"net/http"
	"os"
)

// StartWebServer configures and starts the blocking web server.
func StartWebServer(addr string) {
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

	// Handlers
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		if _, err := w.Write(dashboardHTML); err != nil {
			fmt.Fprintln(os.Stderr, "Error writing response:", err)
		}
	})

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(loginHTML); err != nil {
			fmt.Fprintln(os.Stderr, "Error writing response:", err)
		}
	})

	r.HandleFunc("/logout", handleLogout)

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API routes
	r.HandleFunc("/api/has-password", handleHasPassword)
	r.HandleFunc("/api/login", handleLogin)
	r.HandleFunc("/api/set-password", handleSetPassword)
	r.HandleFunc("/api/search", apiSearch)
	r.HandleFunc("/api/block", apiBlock)
	r.HandleFunc("/api/blocklist", apiBlockList)
	r.HandleFunc("/api/blocklist/clear", apiClearBlocklist)
	r.HandleFunc("/api/blocklist/save", apiSaveBlocklist)
	r.HandleFunc("/api/blocklist/load", apiLoadBlocklist)
	r.HandleFunc("/api/unblock", apiUnblock)

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, authMiddleware(r)); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}
