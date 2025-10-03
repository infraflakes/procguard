package block

import (
	"fmt"
	"procguard/internal/client"

	"github.com/spf13/cobra"
)

// BlockLoadCmd defines the command for loading a blocklist from a file.
var BlockLoadCmd = &cobra.Command{
	Use:   "load <file>",
	Short: "Load block-list from a file, merging with the existing list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		c := client.New()

		if err := c.LoadBlocklist(filePath); err != nil {
			return err
		}

		fmt.Println("block-list loaded and merged from:", filePath)
		return nil
	},
}
