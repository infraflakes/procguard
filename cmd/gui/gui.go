package gui

import (

	_ "embed"
	"encoding/json"
	"fmt"

	"net"
	"net/http"
	"os"
	"os/exec"
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
	exe, _ := os.Executable()
	exec.Command(exe, "daemon").Start()

	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(dashboard)
	})
	r.HandleFunc("/api/search", apiSearch)
	r.HandleFunc("/api/block", apiBlock)
	r.HandleFunc("/api/blocklist", apiBlockList)
	r.HandleFunc("/api/unblock", apiUnblock)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting server:", err)
		os.Exit(1)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	fmt.Println("GUI listening on http://" + addr)
	openBrowser("http://" + addr)

	if err := http.Serve(ln, r); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	exe, err := os.Executable()
	if err != nil {
		http.Error(w, "Could not find executable path", 500)
		return
	}
	out, _ := exec.Command(exe, "find", q).Output()
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
	exe, err := os.Executable()
	if err != nil {
		http.Error(w, "Could not find executable path", 500)
		return
	}
	exec.Command(exe, "block", "add", req.Name).Run()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func apiBlockList(w http.ResponseWriter, r *http.Request) {
	exe, err := os.Executable()
	if err != nil {
		http.Error(w, "Could not find executable path", 500)
		return
	}
	out, _ := exec.Command(exe, "block", "list", "--json").Output()
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
	exe, err := os.Executable()
	if err != nil {
		http.Error(w, "Could not find executable path", 500)
		return
	}
	for _, name := range req.Names {
		exec.Command(exe, "block", "rm", name).Run()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}


