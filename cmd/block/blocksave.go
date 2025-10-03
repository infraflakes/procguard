package block

import (
	"fmt"
	"procguard/internal/client"

	"github.com/spf13/cobra"
)

// BlockSaveCmd defines the command for saving the current blocklist to a file.
var BlockSaveCmd = &cobra.Command{
	Use:   "save <file>",
	Short: "Save current block-list to chosen path (CLI: specify path; GUI: will hook native dialog)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dest := args[0]
		c := client.New()

		if err := c.SaveBlocklist(dest); err != nil {
			return err
		}

		fmt.Println("saved to:", dest)
		return nil
	},
}
