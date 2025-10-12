package api

import (
	"encoding/json"
	"io"
	"net/http"
	"procguard/internal/blocklist"
	"slices"
	"strings"
	"time"
)

func (s *Server) apiBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := blocklist.LoadApp()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		if !slices.Contains(list, lowerName) {
			list = append(list, lowerName)
			// Block the executable file.
			//if err := blocklist.BlockExecutable(lowerName); err != nil {
			//	s.Logger.Printf("Failed to block executable %s: %v", lowerName, err)
			//	// Continue trying to block other executables
			//}
		}
	}

	if err := blocklist.SaveApp(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiUnblock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := blocklist.LoadApp()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		list = slices.DeleteFunc(list, func(item string) bool {
			return item == lowerName
		})
		// Unblock the executable file.
		//if err := blocklist.UnblockExecutable(lowerName); err != nil {
		//	s.Logger.Printf("Failed to unblock executable %s: %v", lowerName, err)
		//	// Continue trying to unblock other executables
		//}
	}

	if err := blocklist.SaveApp(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiBlockList(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.LoadApp()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.Logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiClearBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := blocklist.ClearApp(); err != nil {
		http.Error(w, "Failed to clear blocklist", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) apiSaveBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.LoadApp()
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

func (s *Server) apiLoadBlocklist(w http.ResponseWriter, r *http.Request) {
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

	existingList, err := blocklist.LoadApp()
	if err != nil {
		http.Error(w, "Failed to load existing blocklist", http.StatusInternalServerError)
		return
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	if err := blocklist.SaveApp(existingList); err != nil {
		http.Error(w, "Failed to save merged blocklist", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}