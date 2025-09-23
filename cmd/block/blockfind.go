package block

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var BlockFindCmd = &cobra.Command{
	Use:   "find",
	Short: "List all *.blocked files on PATH and CWD (helps GUI show unblock candidates)",
	Run: func(cmd *cobra.Command, args []string) {
		// scan CWD + PATH dirs
		dirs := []string{"."}
		dirs = append(dirs, filepath.SplitList(os.Getenv("PATH"))...)
		found := 0
		for _, dir := range dirs {
			matches, _ := filepath.Glob(filepath.Join(dir, "*.blocked"))
			for _, m := range matches {
				fmt.Println(strings.TrimSuffix(m, ".blocked"))
				found++
			}
		}
		if found == 0 {
			fmt.Println("no *.blocked files found")
		}
	},
}
