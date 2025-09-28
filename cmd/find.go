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
	RunE:  runFindE,
}

// parseTime is a new helper function to handle our specific time logic.
func parseTime(input string) (time.Time, error) {
	now := time.Now()
	// Check for relative time strings first, case-insensitively
	lowerInput := strings.ToLower(strings.TrimSpace(input))

	switch lowerInput {
	case "now":
		return now, nil
	case "1 hour ago":
		return now.Add(-1 * time.Hour), nil
	case "24 hours ago":
		return now.Add(-24 * time.Hour), nil
	case "7 days ago":
		return now.AddDate(0, 0, -7), nil
	}

	// If it's not a relative string, try parsing it as a timestamp
	// using the original, case-sensitive input string.
	layouts := []string{
		"2006-01-02 15:04:05", // The format we write
		"2006-01-02T15:04",    // The format from datetime-local
		"2006-01-02T15:04:05", // The format from datetime-local with seconds
	}

	for _, layout := range layouts {
		// Use the original 'input' string here, not 'lowerInput'
		parsedTime, err := time.ParseInLocation(layout, input, time.Local)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", input)
}

// runFindE searches the process log file for lines containing the user's query.
func runFindE(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])

	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = parseTime(since)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not parse 'since' time.", err)
		}
	}

	if until != "" {
		untilTime, err = parseTime(until)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not parse 'until' time.", err)
		}
	}

	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("cannot open log: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing file:", err)
		}
	}()

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
		}
	}
	return nil
}
