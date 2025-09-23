package test

import (
	"fmt"
	"os"
	"path/filepath"

	"procguard/cmd/block" // your real package
)

func main() {
	cacheDir, _ := os.UserCacheDir()
	p := filepath.Join(cacheDir, "procguard", "blocklist.json")

	fmt.Println("===  DEBUG  ===")
	fmt.Println("Full path  :", p)
	fmt.Println("Exists?    :", exists(p))

	// try to read **raw**
	b, rawErr := os.ReadFile(p)
	fmt.Printf("Read error : %v\n", rawErr)
	fmt.Printf("Raw content: %q\n", string(b))

	// try your helper
	list, helpErr := block.LoadBlockList()
	fmt.Printf("Helper err : %v\n", helpErr)
	fmt.Printf("List       : %#v\n", list)
	fmt.Println("=== END DEBUG ===")
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
