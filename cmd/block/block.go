package block

import (
	"github.com/spf13/cobra"
)

func init() {
	// register sub-commands
	BlockCmd.AddCommand(BlockAddCmd, BlockRmCmd, BlockListCmd, BlockClearCmd, BlockSaveCmd, BlockFindCmd, BlockLoadCmd)
	BlockCmd.PersistentFlags().String("token", "", "Auth token for GUI calls")
}

var BlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage block-list (add, remove, list, clear, save, find-blocked)",
}
