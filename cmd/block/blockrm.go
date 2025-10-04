package block

import (
	"procguard/internal/client"

	"github.com/spf13/cobra"
)

func init() {
	BlockRmCmd.Flags().Bool("json", false, "output json for gui")
}

var BlockRmCmd = &cobra.Command{
	Use:   "rm <exe>",
	Short: "Remove program from block-list and unblock executable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		c := client.New()

		if err := c.Unblock(name); err != nil {
			return err
		}

		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "removed", name)
		return nil
	},
}
