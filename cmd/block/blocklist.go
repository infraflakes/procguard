package block

import (
	"github.com/spf13/cobra"
)

func init() {
	BlockListCmd.Flags().Bool("json", false, "output array for gui")
}

// BlockListCmd defines the command for displaying the current blocklist.
var BlockListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current block-list",
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := LoadBlockList()
		isJSON, _ := cmd.Flags().GetBool("json")
		// Use the centralized ReplyList function to handle the output.
		ReplyList(isJSON, list)
	},
}
