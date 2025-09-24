package gui

import (

	_ "embed"
	"encoding/json"
	"fmt"


	"net/http"
	"os"

	"path/filepath"
	"strings"


	"github.com/spf13/cobra"
)

//go:embed dashboard.html
var dashboard []byte

var logPath string

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
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(dashboard)
	})
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/api/search", apiSearch)
	r.HandleFunc("/api/block", apiBlock)
	r.HandleFunc("/api/blocklist", apiBlockList)
	r.HandleFunc("/api/unblock", apiUnblock)

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	cmd, err := runProcGuardCommand("find", q)
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
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cmd, err := runProcGuardCommand("block", "add", req.Name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	cmd.Run()
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


