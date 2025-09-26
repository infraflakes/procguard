package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	since string
	until string
)

func init() {
	findCmd.Flags().StringVar(&since, "since", "", "Show logs since a specific time (e.g., '1 hour ago', '2025-09-26 14:00:00')")
	findCmd.Flags().StringVar(&until, "until", "", "Show logs until a specific time (e.g., 'now')")
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find <name>",
	Short: "Find log lines by program name (case-insensitive)",
	Args:  cobra.ExactArgs(1),
	Run:   runFind,
}

// parseTime is a new helper function to handle our specific time logic.
func parseTime(input string) (time.Time, error) {
	now := time.Now()
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "now":
		return now, nil
	case "1 hour ago":
		return now.Add(-1 * time.Hour), nil
	case "24 hours ago":
		return now.Add(-24 * time.Hour), nil
	case "7 days ago":
		return now.AddDate(0, 0, -7), nil
	}

	// Fallback to parsing a specific timestamp
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", input, time.Local)
	if err == nil {
		return parsedTime, nil
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", input)
}

// runFind searches the process log file for lines containing the user's query.
func runFind(cmd *cobra.Command, args []string) {
	query := strings.ToLower(args[0])

	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = parseTime(since)
		if err != nil {
			// Silently ignore parse errors for now, or we could log to stderr
		}
	}

	if until != "" {
		untilTime, err = parseTime(until)
		if err != nil {
			// Silently ignore
		}
	}

	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	file, err := os.Open(logFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot open log:", err)
		os.Exit(1)
	}
	defer file.Close()

	found := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " | ")
		if len(parts) < 4 {
			continue
		}

		logTime, err := time.ParseInLocation("2006-01-02 15:04:05", parts[0], time.Local)
		if err != nil {
			continue
		}

		if !sinceTime.IsZero() && logTime.Before(sinceTime) {
			continue
		}
		if !untilTime.IsZero() && logTime.After(untilTime) {
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
