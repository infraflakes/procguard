package block

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// BlockFindCmd defines the command for finding files that have been blocked by ProcGuard.
// This is particularly useful on Windows, where blocking is done by renaming the file.
var BlockFindCmd = &cobra.Command{
	Use:   "find",
	Short: "List all *.blocked files on PATH and CWD (helps GUI show unblock candidates)",
	Run: func(cmd *cobra.Command, args []string) {
		// Search for blocked files in the current working directory and all directories in the system's PATH.
		dirs := []string{"."}
		dirs = append(dirs, filepath.SplitList(os.Getenv("PATH"))...)
		found := 0
		for _, dir := range dirs {
			matches, _ := filepath.Glob(filepath.Join(dir, "*.blocked"))
			for _, m := range matches {
				// Print the original name of the blocked file.
				fmt.Println(strings.TrimSuffix(m, ".blocked"))
				found++
			}
		}
		if found == 0 {
			fmt.Println("no *.blocked files found")
		}
	},
}
