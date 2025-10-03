package block

import (
	"procguard/internal/client"

	"github.com/spf13/cobra"
)

func init() {
	BlockListCmd.Flags().Bool("json", false, "output array for gui")
}

// BlockListCmd defines the command for displaying the current blocklist.
var BlockListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current block-list",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New()
		list, err := c.GetBlocklist()
		if err != nil {
			return err
		}

		isJSON, _ := cmd.Flags().GetBool("json")
		ReplyList(isJSON, list)
		return nil
	},
}