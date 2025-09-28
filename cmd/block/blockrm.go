package block

import (
	"fmt"
	"procguard/internal/blocklist"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	BlockRmCmd.Flags().Bool("json", false, "output json for gui")
}

var BlockRmCmd = &cobra.Command{
	Use:   "rm <exe>",
	Short: "Remove program from block-list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// All entries are stored in lowercase for case-insensitive matching.
		base := strings.ToLower(args[0])
		list, err := blocklist.Load()
		if err != nil {
			return err
		}

		// Find the index of the program to be removed.
		idx := slices.Index(list, base)
		if idx == -1 {
			isJSON, _ := cmd.Flags().GetBool("json")
			Reply(isJSON, "not found", base)
			return nil
		}

		// Remove the element from the list.
		list = slices.Delete(list, idx, idx+1)
		if err := blocklist.Save(list); err != nil {
			return fmt.Errorf("save: %w", err)
		}

		// Respond to the user with the result of the operation.
		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "removed", base)
		return nil
	},
}
