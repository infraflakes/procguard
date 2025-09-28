package gui

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/internal/auth"
	"procguard/internal/config"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

//go:embed dashboard.html
var dashboardHTML []byte

//go:embed login.html
var loginHTML []byte

var logPath string
var isAuthenticated bool
var mu sync.Mutex

// runProcGuardCommand executes the main procguard binary with the given arguments.
// This is a workaround to avoid circular dependencies and to interact with the CLI
// commands from the GUI server.
func runProcGuardCommand(args ...string) (*exec.Cmd, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to find executable path: %w", err)
	}
	cmd := exec.Command(executable, args...)
	cmd.Stderr = new(bytes.Buffer) // Capture stderr
	return cmd, nil
}

var GuiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Run the web-based GUI",
	Run:   runGUI,
}

func init() {
	cacheDir, _ := os.UserCacheDir()
	logPath = filepath.Join(cacheDir, "procguard", "events.log")
}

func runGUI(cmd *cobra.Command, args []string) {
	const defaultPort = "58141"
	addr := "127.0.0.1:" + defaultPort
	fmt.Println("Starting GUI on http://" + addr)
	StartWebServer(addr)
}

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
		w.Write(dashboardHTML)
	})

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(loginHTML)
	})

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
	r.HandleFunc("/api/unblock", apiUnblock)

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, authMiddleware(r)); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

func handleHasPassword(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}
	hasPassword := cfg.PasswordHash != ""
	json.NewEncoder(w).Encode(map[string]bool{"hasPassword": hasPassword})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
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

	if auth.CheckPasswordHash(req.Password, cfg.PasswordHash) {
		mu.Lock()
		isAuthenticated = true
		mu.Unlock()
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	} else {
		json.NewEncoder(w).Encode(map[string]bool{"success": false})
	}
}

func handleSetPassword(w http.ResponseWriter, r *http.Request) {
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

	mu.Lock()
	isAuthenticated = true
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	since := r.URL.Query().Get("since")
	until := r.URL.Query().Get("until")

	args := []string{"find", q}
	if since != "" {
		args = append(args, "--since", since)
	}
	if until != "" {
		args = append(args, "--until", until)
	}

	cmd, err := runProcGuardCommand(args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, _ := cmd.Output()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var jsonLines [][]string
	for _, l := range lines {
		if l != "" {
			jsonLines = append(jsonLines, strings.Split(l, " | "))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonLines)
}

func apiBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, name := range req.Names {
		cmd, err := runProcGuardCommand("block", "add", name)
		if err != nil {
			// Decide if you want to stop or continue on error
			continue
		}
		cmd.Run()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func apiBlockList(w http.ResponseWriter, r *http.Request) {
	cmd, err := runProcGuardCommand("block", "list", "--json")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	out, _ := cmd.Output()
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func apiUnblock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, name := range req.Names {
		cmd, err := runProcGuardCommand("block", "rm", name)
		if err != nil {
			// Decide if you want to stop or continue on error
			continue
		}
		cmd.Run()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
