package gui

import (
	"encoding/json"
	"net/http"
	"procguard/internal/logsearch"
	"strings"
)

func (s *Server) apiSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	sinceStr := r.URL.Query().Get("since")
	untilStr := r.URL.Query().Get("until")

	results, err := logsearch.Search(query, sinceStr, untilStr)
	if err != nil {
		http.Error(w, "Failed to search logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}
