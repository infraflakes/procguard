package blocklist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const blockListFile = "blocklist.json"

// Load reads the blocklist file from the user's cache directory,
// normalizes all entries to lowercase, and returns them as a slice of strings.
// If the file doesn't exist, it returns an empty list.
func Load() ([]string, error) {
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
	_ = json.Unmarshal(b, &list)

	// Normalize all entries to lowercase for case-insensitive comparison.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// Save writes the given list of strings to the blocklist file in the
// user's cache directory. It normalizes all entries to lowercase before saving.
// It also sets appropriate file permissions to secure the file.
func Save(list []string) error {
	// Normalize all entries to lowercase to ensure consistency.
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, _ := os.UserCacheDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "procguard"), 0755)
	p := filepath.Join(cacheDir, "procguard", blockListFile)

	// Marshal the list to JSON with indentation for readability.
	b, _ := json.MarshalIndent(list, "", "  ")
	if err := os.WriteFile(p, b, 0600); err != nil {
		return err
	}

	// Apply platform-specific file locking to prevent unauthorized modification.
	return platformLock(p) // build-tag dispatch
}
