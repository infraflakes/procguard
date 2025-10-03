package block

import (
	"fmt"
	"procguard/internal/blocklist"
	"procguard/internal/platform"

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
		status, err := blocklist.Add(name)
		if err != nil {
			return err
		}

		if status == "exists" {
			isJSON, _ := cmd.Flags().GetBool("json")
			Reply(isJSON, "exists", name)
			return nil
		}

		if err := platform.BlockExecutable(name); err != nil {
			// If blocking fails, we should probably roll back adding to the list,
			// but for now, we'll just return the error.
			return fmt.Errorf("failed to block executable: %w", err)
		}

		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "added", name)
		return nil
	},
}
