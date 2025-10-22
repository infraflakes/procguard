package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const blockListFile = "blocklist.json"

type AppDetails struct {
	Name    string `json:"name"`
	ExePath string `json:"exe_path"`
}

// LoadAppWithDetails loads the blocklist and enriches it with the latest exe_path from the database.
func LoadAppWithDetails(db *sql.DB) ([]AppDetails, error) {
	names, err := LoadApp()
	if err != nil {
		return nil, fmt.Errorf("could not load app blocklist names: %w", err)
	}

	if len(names) == 0 {
		return []AppDetails{}, nil
	}

	details := make([]AppDetails, 0, len(names))
	for _, name := range names {
		var exePath string
		// Find the most recent exe_path for the given process name.
		err := db.QueryRow("SELECT exe_path FROM app_events WHERE process_name = ? AND exe_path IS NOT NULL ORDER BY start_time DESC LIMIT 1", name).Scan(&exePath)
		if err != nil {
			if err == sql.ErrNoRows {
				// If no path is found, we can still return the name.
				exePath = ""
			} else {
				// For other errors, log them but continue.
				GetLogger().Printf("Error querying exe_path for %s: %v", name, err)
				exePath = ""
			}
		}
		details = append(details, AppDetails{Name: name, ExePath: exePath})
	}

	return details, nil
}

// LoadApp reads the blocklist file from the user's cache directory,
// normalizes all entries to lowercase, and returns them as a slice of strings.
// If the file doesn't exist, it returns an empty list.
func LoadApp() ([]string, error) {
	cacheDir, _ := os.UserCacheDir()
	p := filepath.Join(cacheDir, "procguard", blockListFile)

	// If the blocklist file doesn't exist, return an empty list.
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocklist: %w", err)
	}

	// Normalize all entries to lowercase for case-insensitive comparison.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// SaveApp writes the given list of strings to the blocklist file in the
// user's cache directory. It normalizes all entries to lowercase before saving.
// It also sets appropriate file permissions to secure the file.
func SaveApp(list []string) error {
	// Normalize all entries to lowercase to ensure consistency.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, _ := os.UserCacheDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "procguard"), 0755)
	p := filepath.Join(cacheDir, "procguard", blockListFile)

	// Marshal the list to JSON with indentation for readability.
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		return err
	}

	// Apply platform-specific file locking to prevent unauthorized modification.
	return platformLock(p) // build-tag dispatch
}

// AddApp adds a program to the blocklist.
func AddApp(name string) (string, error) {
	list, err := LoadApp()
	if err != nil {
		return "", err
	}

	lowerName := strings.ToLower(name)
	for _, v := range list {
		if v == lowerName {
			return "exists", nil
		}
	}

	list = append(list, lowerName)
	if err := SaveApp(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "added", nil
}

// RemoveApp removes a program from the blocklist.
func RemoveApp(name string) (string, error) {
	list, err := LoadApp()
	if err != nil {
		return "", err
	}

	lowerName := strings.ToLower(name)
	idx := slices.Index(list, lowerName)
	if idx == -1 {
		return "not found", nil
	}

	list = slices.Delete(list, idx, idx+1)
	if err := SaveApp(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "removed", nil
}

// ClearApp clears the blocklist.
func ClearApp() error {
	return SaveApp([]string{})
}

// SaveAppToFile saves the current blocklist to a file.
func SaveAppToFile(path string) error {
	list, err := LoadApp()
	if err != nil {
		return err
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

// LoadAppFromFile loads a blocklist from a file, merging it with the existing list.
func LoadAppFromFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}

	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}

	err = json.Unmarshal(content, &newEntries)
	if err != nil {
		err2 := json.Unmarshal(content, &savedList)
		if err2 != nil {
			return fmt.Errorf("load: invalid JSON format in %s", path)
		}
		newEntries = savedList.Blocked
	}

	existingList, err := LoadApp()
	if err != nil {
		return err
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	return SaveApp(existingList)
}
