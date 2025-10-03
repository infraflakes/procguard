package block

import (
	"fmt"
	"procguard/internal/blocklist"
	"procguard/internal/platform"

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
		status, err := blocklist.Remove(name)
		if err != nil {
			return err
		}

		if status == "not found" {
			isJSON, _ := cmd.Flags().GetBool("json")
			Reply(isJSON, "not found", name)
			return nil
		}

		if err := platform.UnblockExecutable(name); err != nil {
			return fmt.Errorf("failed to unblock executable: %w", err)
		}

		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "removed", name)
		return nil
	},
}
