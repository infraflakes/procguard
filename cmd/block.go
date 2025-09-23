package cmd

import (
	"procguard/cmd/block"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(blockCmd)
	// register sub-commands
	blockCmd.AddCommand(block.BlockAddCmd, block.BlockRmCmd, block.BlockListCmd, block.BlockClearCmd, block.BlockSaveCmd, block.BlockFindCmd, block.BlockLoadCmd)
}

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage block-list (add, remove, list, clear, save, find-blocked)",
}
