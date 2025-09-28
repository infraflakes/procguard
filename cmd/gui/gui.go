package gui

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"procguard/internal/auth"
	"procguard/internal/blocklist"
	"procguard/internal/config"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

//go:embed dashboard.html
var dashboardHTML []byte

//go:embed login.html
var loginHTML []byte

var logPath string
var isAuthenticated bool
var mu sync.Mutex

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

func handleLogout(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	isAuthenticated = false
	mu.Unlock()
	http.Redirect(w, r, "/login", http.StatusFound)
}

func apiSaveBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to get blocklist", http.StatusInternalServerError)
		return
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=procguard_blocklist.json")
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func apiLoadBlocklist(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusInternalServerError)
		return
	}

	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}

	err = json.Unmarshal(content, &newEntries)
	if err != nil {
		err2 := json.Unmarshal(content, &savedList)
		if err2 != nil {
			http.Error(w, "Invalid JSON format in uploaded file", http.StatusBadRequest)
			return
		}
		newEntries = savedList.Blocked
	}

	existingList, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load existing blocklist", http.StatusInternalServerError)
		return
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	if err := blocklist.Save(existingList); err != nil {
		http.Error(w, "Failed to save merged blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func apiClearBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := blocklist.Save([]string{}); err != nil {
		http.Error(w, "Failed to clear blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
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
	
		json.NewEncoder(w).Encode(map[string]bool{"success": true})}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	sinceStr := r.URL.Query().Get("since")
	untilStr := r.URL.Query().Get("until")

	var sinceTime, untilTime time.Time
	var err error

	if sinceStr != "" {
		sinceTime, err = parseTime(sinceStr)
		if err != nil {
			http.Error(w, "Invalid 'since' time format", http.StatusBadRequest)
			return
		}
	}

	if untilStr != "" {
		untilTime, err = parseTime(untilStr)
		if err != nil {
			http.Error(w, "Invalid 'until' time format", http.StatusBadRequest)
			return
		}
	}

	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	file, err := os.Open(logFile)
	if err != nil {
		http.Error(w, "Cannot open log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var results [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " | ")
		if len(parts) < 4 {
			continue
		}

		logTime, err := time.ParseInLocation("2006-01-02 15:04:05", parts[0], time.Local)
		if err != nil {
			continue
		}

		if !sinceTime.IsZero() && logTime.Before(sinceTime) {
			continue
		}
		if !untilTime.IsZero() && logTime.After(untilTime) {
			continue
		}

		exe := strings.ToLower(parts[1])
		parentExe := strings.ToLower(parts[3])
		if strings.Contains(exe, query) || strings.Contains(parentExe, query) {
			results = append(results, parts)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// parseTime is a helper function to handle our specific time logic for the GUI.
func parseTime(input string) (time.Time, error) {
	now := time.Now()
	lowerInput := strings.ToLower(strings.TrimSpace(input))

	switch lowerInput {
	case "now":
		return now, nil
	case "1 hour ago":
		return now.Add(-1 * time.Hour), nil
	case "24 hours ago":
		return now.Add(-24 * time.Hour), nil
	case "7 days ago":
		return now.AddDate(0, 0, -7), nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		parsedTime, err := time.ParseInLocation(layout, input, time.Local)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", input)
}


func apiBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		if !slices.Contains(list, lowerName) {
			list = append(list, lowerName)
		}
	}

	if err := blocklist.Save(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func apiBlockList(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func apiUnblock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		list = slices.DeleteFunc(list, func(item string) bool {
			return item == lowerName
		})
	}

	if err := blocklist.Save(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
