package block

import (
	"procguard/internal/client"

	"github.com/spf13/cobra"
)

func init() {
	BlockAddCmd.Flags().Bool("json", false, "output json for gui")
}

var BlockAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add program to block-list and block executable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		c := client.New()

		if err := c.Block(name); err != nil {
			return err
		}

		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "added", name)
		return nil
	},
}
