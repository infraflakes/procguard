package cmd

import (
	"fmt"

	"procguard/internal/database"
	"procguard/internal/logsearch"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenDB() // Use OpenDB for read-only clients
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		query := strings.ToLower(args[0])

		results, err := logsearch.Search(db, query, since, until)
		if err != nil {
			return err
		}

		for _, parts := range results {
			fmt.Println(strings.Join(parts, " | "))
		}

		return nil
	},
}
