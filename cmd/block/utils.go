package block

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const blockListFile = "blocklist.json"

// ----------  exported helpers  ----------

// LoadBlockList reads the blocklist file from the user's cache directory,
// normalizes all entries to lowercase, and returns them as a slice of strings.
// If the file doesn't exist, it returns an empty list.
func LoadBlockList() ([]string, error) {
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

// SaveBlockList writes the given list of strings to the blocklist file in the
// user's cache directory. It normalizes all entries to lowercase before saving.
// It also sets appropriate file permissions to secure the file.
func SaveBlockList(list []string) error {
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

// Reply provides a standardized way to respond to CLI commands, supporting both
// plain text and JSON output formats.
func Reply(isJSON bool, status, exe string) {
	if isJSON {
		out, _ := json.Marshal(map[string]string{"status": status, "exe": exe})
		fmt.Println(string(out))
	} else {
		fmt.Println(status+":", exe)
	}
}

// ReplyList provides a standardized way to output a list of strings, supporting
// both plain text and JSON formats.
func ReplyList(isJSON bool, list []string) {
	if isJSON {
		out, _ := json.Marshal(list)
		fmt.Println(string(out))
	} else {
		if len(list) == 0 {
			fmt.Println("block-list is empty")
			return
		}
		fmt.Println("blocked programs:")
		for _, v := range list {
			fmt.Println(" -", v)
		}
	}
}
