package api

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleGetWebLogs(w http.ResponseWriter, r *http.Request) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		http.Error(w, "Failed to get user cache dir", http.StatusInternalServerError)
		return
	}
	logPath := filepath.Join(cacheDir, "procguard", "web-logs", "events.log")

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([][]string{})
			return
		}
		http.Error(w, "Failed to open web log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var entries [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "URL:"); idx != -1 {
			timestamp := strings.TrimSpace(line[:idx])
			url := strings.TrimSpace(line[idx+4:])
			if url != "" {
				entries = append(entries, []string{timestamp, url})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}
