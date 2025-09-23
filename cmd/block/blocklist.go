package block

import (

	"github.com/spf13/cobra"
)

func init() {
	BlockListCmd.Flags().Bool("json", false, "output array for gui")
}

var BlockListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current block-list",
	Run: func(cmd *cobra.Command, args []string) {
		list, _ := LoadBlockList()
		isJSON, _ := cmd.Flags().GetBool("json")
		ReplyList(isJSON, list)
	},
}
