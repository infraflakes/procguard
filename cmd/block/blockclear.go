package block

import (
	"fmt"
	"os"
	"procguard/internal/blocklist"

	"github.com/spf13/cobra"
)

var BlockClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Empty the block-list completely",
	Run: func(cmd *cobra.Command, args []string) {
		// This is done by saving an empty list to the blocklist file.
		if err := blocklist.Save([]string{}); err != nil {
			fmt.Fprintln(os.Stderr, "clear:", err)
			os.Exit(1)
		}
		fmt.Println("block-list cleared")
	},
}
