package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"procguard/pkg/blocklist/web"
	"slices"
	"strings"
	"time"
)

func (s *Server) handleGetWebBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := web.Load()
	if err != nil {
		http.Error(w, "Failed to load web blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleAddWebBlocklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := web.Add(req.Domain); err != nil {
		http.Error(w, "Failed to add to web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleRemoveWebBlocklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := web.Remove(req.Domain); err != nil {
		http.Error(w, "Failed to remove from web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleClearWebBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := web.Save([]string{}); err != nil {
		http.Error(w, "Failed to clear web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleSaveWebBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := web.Load()
	if err != nil {
		http.Error(w, "Failed to get web blocklist", http.StatusInternalServerError)
		return
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=procguard_web_blocklist.json")
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		s.logger.Printf("Error writing response: %v", err)
	}
}

func (s *Server) handleLoadWebBlocklist(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Printf("Error closing file: %v", err)
		}
	}()

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

	existingList, err := web.Load()
	if err != nil {
		http.Error(w, "Failed to load existing web blocklist", http.StatusInternalServerError)
		return
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	if err := web.Save(existingList); err != nil {
		http.Error(w, "Failed to save merged web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleGetWebLogs(w http.ResponseWriter, r *http.Request) {
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

	q := "SELECT url, timestamp FROM web_events WHERE 1=1"
	args := make([]interface{}, 0)

	if !sinceTime.IsZero() {
		q += " AND timestamp >= ?"
		args = append(args, sinceTime.Unix())
	}

	if !untilTime.IsZero() {
		q += " AND timestamp <= ?"
		args = append(args, untilTime.Unix())
	}

	q += " ORDER BY timestamp DESC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		http.Error(w, "Failed to query web logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var entries [][]string
	for rows.Next() {
		var url string
		var timestamp int64
		if err := rows.Scan(&url, &timestamp); err != nil {
			continue
		}
		timestampStr := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		entries = append(entries, []string{timestampStr, url})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

// parseTime is a helper function to handle our specific time logic.
// This is duplicated from the logsearch package for simplicity for now.
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
