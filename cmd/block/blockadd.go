package block

import (
	"fmt"
	"procguard/internal/blocklist"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	BlockAddCmd.Flags().Bool("json", false, "output json for gui")
}

var BlockAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add program to block-list (OS-agnostic)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// All entries are stored in lowercase for case-insensitive matching.
		name := strings.ToLower(args[0])
		list, err := blocklist.Load()
		if err != nil {
			return err
		}

		// Check if the program is already in the blocklist.
		for _, v := range list {
			if v == name {
				isJSON, _ := cmd.Flags().GetBool("json")
				Reply(isJSON, "exists", name)
				return nil
			}
		}

		// Add the new program to the list and save it.
		list = append(list, name)
		if err := blocklist.Save(list); err != nil {
			return fmt.Errorf("save: %w", err)
		}

		// Respond to the user with the result of the operation.
		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "added", name)
		return nil
	},
}
