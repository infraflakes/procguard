package api

import (
	"encoding/json"
	"io"
	"net/http"
	"procguard/internal/blocklist"
	"procguard/internal/util"
	"slices"
	"time"
)

func (s *Server) handleGetWebBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.LoadWeb()
	if err != nil {
		http.Error(w, "Failed to load web blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
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

	if _, err := blocklist.AddWeb(req.Domain); err != nil {
		http.Error(w, "Failed to add to web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
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

	if _, err := blocklist.RemoveWeb(req.Domain); err != nil {
		http.Error(w, "Failed to remove from web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleClearWebBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := blocklist.SaveWeb([]string{}); err != nil {
		http.Error(w, "Failed to clear web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleSaveWebBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.LoadWeb()
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
		s.Logger.Printf("Error writing response: %v", err)
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
			s.Logger.Printf("Error closing file: %v", err)
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

	existingList, err := blocklist.LoadWeb()
	if err != nil {
		http.Error(w, "Failed to load existing web blocklist", http.StatusInternalServerError)
		return
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	if err := blocklist.SaveWeb(existingList); err != nil {
		http.Error(w, "Failed to save merged web blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleGetWebLogs(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")
	untilStr := r.URL.Query().Get("until")

	var sinceTime, untilTime time.Time
	var err error

	if sinceStr != "" {
		sinceTime, err = util.ParseTime(sinceStr)
		if err != nil {
			http.Error(w, "Invalid 'since' time format", http.StatusBadRequest)
			return
		}
	}

	if untilStr != "" {
		untilTime, err = util.ParseTime(untilStr)
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
		s.Logger.Printf("Error encoding response: %v", err)
	}
}
