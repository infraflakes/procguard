package api

import (
	"encoding/json"
	"io"
	"net/http"
	"procguard/internal/data"
	"slices"
	"strings"
	"time"
)

// handleBlockApps adds one or more applications to the blocklist.
// It expects a JSON request with a `names` field containing a list of application names.
func (s *Server) handleBlockApps(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := data.LoadAppBlocklist()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		if !slices.Contains(list, lowerName) {
			list = append(list, lowerName)
		}
	}

	if err := data.SaveAppBlocklist(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleUnblockApps removes one or more applications from the blocklist.
// It expects a JSON request with a `names` field containing a list of application names.
func (s *Server) handleUnblockApps(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := data.LoadAppBlocklist()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		list = slices.DeleteFunc(list, func(item string) bool {
			return item == lowerName
		})
	}

	if err := data.SaveAppBlocklist(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleGetAppBlocklist returns the list of blocked applications with their details.
func (s *Server) handleGetAppBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := data.GetBlockedAppsWithDetails(s.db)
	if err != nil {
		http.Error(w, "Failed to load blocklist with details", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

// handleClearAppBlocklist removes all applications from the blocklist.
func (s *Server) handleClearAppBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := data.ClearAppBlocklist(); err != nil {
		http.Error(w, "Failed to clear blocklist", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handleSaveAppBlocklist saves the current application blocklist to a file for export.
func (s *Server) handleSaveAppBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := data.LoadAppBlocklist()
	if err != nil {
		http.Error(w, "Failed to get blocklist", http.StatusInternalServerError)
		return
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=procguard_blocklist.json")
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		s.Logger.Printf("Error writing response: %v", err)
	}
}

// handleLoadAppBlocklist loads an application blocklist from an uploaded file and merges it with the existing blocklist.
func (s *Server) handleLoadAppBlocklist(w http.ResponseWriter, r *http.Request) {
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

	existingList, err := data.LoadAppBlocklist()
	if err != nil {
		http.Error(w, "Failed to load existing blocklist", http.StatusInternalServerError)
		return
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	if err := data.SaveAppBlocklist(existingList); err != nil {
		http.Error(w, "Failed to save merged blocklist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
