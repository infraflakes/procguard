package gui

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	r := http.NewServeMux()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(dashboard)
	})
	r.HandleFunc("/api/search", apiSearch)
	r.HandleFunc("/api/block", apiBlock)
	r.HandleFunc("/api/events", apiEventStream)

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
	out, _ := exec.Command("procguard", "find", q).Output()
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
	exec.Command("procguard", "block", "add", req.Name).Run()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func apiEventStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			f, err := os.Open(logPath)
			if err != nil {
				// Log file might not exist yet
				continue
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			var buf []string
			for scanner.Scan() {
				buf = append(buf, scanner.Text())
			}

			if len(buf) > 200 {
				buf = buf[len(buf)-200:]
			}

			for _, line := range buf {
				fmt.Fprintf(w, "data: %s\n\n", line)
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}
