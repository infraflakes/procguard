package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(findCmd) }

var findCmd = &cobra.Command{
	Use:   "find <name>",
	Short: "Find log lines by program name (case-insensitive)",
	Args:  cobra.ExactArgs(1),
	Run:   runFind,
}

func runFind(cmd *cobra.Command, args []string) {
	query := strings.ToLower(args[0])
	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	file, err := os.Open(logFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot open log:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	found := 0
	for scanner.Scan() {
		line := scanner.Text()
		// line = date and time | exe | pid | parent_exe
		parts := strings.Split(line, " | ")
		if len(parts) < 4 {
			continue
		}
		exe := strings.ToLower(parts[1])
		parentExe := strings.ToLower(parts[3])
		if strings.Contains(exe, query) || strings.Contains(parentExe, query) {
			fmt.Println(line)
			found++
		}
	}
	if found == 0 {
		fmt.Println("no match for:", args[0])
	}
}
