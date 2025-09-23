package block

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var BlockClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Empty the block-list completely",
	Run: func(cmd *cobra.Command, args []string) {
		if err := SaveBlockList([]string{}); err != nil {
			fmt.Fprintln(os.Stderr, "clear:", err)
			os.Exit(1)
		}
		fmt.Println("block-list cleared")
	},
}
