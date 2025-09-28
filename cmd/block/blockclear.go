package block

import (
	"fmt"
	"procguard/internal/blocklist"

	"github.com/spf13/cobra"
)

var BlockClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Empty the block-list completely",
	RunE: func(cmd *cobra.Command, args []string) error {
		// This is done by saving an empty list to the blocklist file.
		if err := blocklist.Save([]string{}); err != nil {
			return fmt.Errorf("clear: %w", err)
		}
		fmt.Println("block-list cleared")
		return nil
	},
}
