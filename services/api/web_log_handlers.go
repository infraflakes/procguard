package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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
