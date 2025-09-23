package block

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const blockListFile = "blocklist.json"

// ----------  exported helpers  ----------

// LoadBlockList reads & normalises the JSON list.
func LoadBlockList() ([]string, error) {
	cacheDir, _ := os.UserCacheDir()
	p := filepath.Join(cacheDir, "procguard", blockListFile)
	b, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []string
	_ = json.Unmarshal(b, &list)
	// lower-case for case-insensitive compare
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}
	return list, nil
}

// SaveBlockList writes list (pretty) and sets 0600 on Unix, ACL on Win.
func SaveBlockList(list []string) error {
	// lower-case for case-insensitive compare
	for i := range list {
		list[i] = strings.ToLower(list[i])
	}

	cacheDir, _ := os.UserCacheDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "procguard"), 0755)
	p := filepath.Join(cacheDir, "procguard", blockListFile)

	b, _ := json.MarshalIndent(list, "", "  ")
	if err := os.WriteFile(p, b, 0600); err != nil {
		return err
	}
	return platformLock(p) // build-tag dispatch
}

// Reply prints plain text or JSON depending on isJSON flag
func Reply(isJSON bool, status, exe string) {
	if isJSON {
		out, _ := json.Marshal(map[string]string{"status": status, "exe": exe})
		fmt.Println(string(out))
	} else {
		fmt.Println(status+":", exe)
	}
}

// ReplyList prints plain text or JSON depending on isJSON flag
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


// ----------  OS-agnostic block/unblock  ----------

// BlockExecutable blocks **whatever the OS considers runnable**.
func BlockExecutable(name string) error {
	path, err := findExecutable(name)
	if err != nil {
		return err
	}
	return blockFile(path) // build-tag dispatch
}

// UnblockExecutable reverses the above.
func UnblockExecutable(name string) error {
	path, err := findExecutable(name)
	if err != nil {
		return err
	}
	return unblockFile(path) // build-tag dispatch
}

// ----------  internal helpers  ----------

// findExecutable resolves name via PATH or CWD; returns **absolute** path.
func findExecutable(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}
	// allow both “chrome” and “chrome.exe” on Windows
	return exec.LookPath(name)
}
