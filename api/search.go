package api

import (
	"encoding/json"
	"net/http"
	"procguard/internal/data"
	"strings"
)

// handleSearch handles searches for application events.
// It accepts the following query parameters:
// - q: the search query string
// - since: the start of the time range (e.g., "1 hour ago")
// - until: the end of the time range (e.g., "now")
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	sinceStr := r.URL.Query().Get("since")
	untilStr := r.URL.Query().Get("until")

	results, err := data.SearchAppEvents(s.db, query, sinceStr, untilStr)
	if err != nil {
		s.Logger.Printf("Error searching logs: %v", err)
		http.Error(w, "Failed to search logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}
