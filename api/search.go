package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/data"
	"strings"
)

func (s *Server) apiSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	sinceStr := r.URL.Query().Get("since")
	untilStr := r.URL.Query().Get("until")

	s.Logger.Printf("Searching logs with query: '%s', since: '%s', until: '%s'", query, sinceStr, untilStr)

	results, err := data.Search(s.db, query, sinceStr, untilStr)
	if err != nil {
		s.Logger.Printf("Error searching logs: %v", err)
		http.Error(w, "Failed to search logs", http.StatusInternalServerError)
		return
	}

	s.Logger.Printf("Found %d log entries.", len(results))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}
