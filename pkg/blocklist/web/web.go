package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const webBlocklistFile = "web_blocklist.json"

// Load reads the web blocklist file from the user's cache directory,
// normalizes all entries to lowercase, and returns them as a slice of strings.
// If the file doesn't exist, it returns an empty list.
func Load() ([]string, error) {
	cacheDir, _ := os.UserCacheDir()
	p := filepath.Join(cacheDir, "procguard", webBlocklistFile)

	// If the blocklist file doesn't exist, return an empty list.
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var list []string
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal web blocklist: %w", err)
	}

	// Normalize all entries to lowercase for case-insensitive comparison.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// Save writes the given list of strings to the web blocklist file in the
// user's cache directory. It normalizes all entries to lowercase before saving.
func Save(list []string) error {
	// Normalize all entries to lowercase to ensure consistency.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, _ := os.UserCacheDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "procguard"), 0755)
	p := filepath.Join(cacheDir, "procguard", webBlocklistFile)

	// Marshal the list to JSON with indentation for readability.
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal web blocklist: %w", err)
	}
	return os.WriteFile(p, b, 0600)
}

// Add adds a domain to the web blocklist.
func Add(domain string) (string, error) {
	list, err := Load()
	if err != nil {
		return "", err
	}

	lowerDomain := strings.ToLower(domain)
	for _, v := range list {
		if v == lowerDomain {
			return "exists", nil
		}
	}

	list = append(list, lowerDomain)
	if err := Save(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "added", nil
}

// Remove removes a domain from the web blocklist.
func Remove(domain string) (string, error) {
	list, err := Load()
	if err != nil {
		return "", err
	}

	lowerDomain := strings.ToLower(domain)
	idx := slices.Index(list, lowerDomain)
	if idx == -1 {
		return "not found", nil
	}

	list = slices.Delete(list, idx, idx+1)
	if err := Save(list); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	return "removed", nil
}

// Clear clears the web blocklist.
func Clear() error {
	return Save([]string{})
}
