package block

import (
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		name := strings.ToLower(args[0])
		list, _ := LoadBlockList()

		// normalised inside LoadBlockList
		for _, v := range list {
			if v == name {
				isJSON, _ := cmd.Flags().GetBool("json")
				Reply(isJSON, "exists", name)
				return
			}
		}

		list = append(list, name)
		if err := SaveBlockList(list); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}
		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "added", name)
	},
}
