package block

import (
	"fmt"
	"os"
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
	Run: func(cmd *cobra.Command, args []string) {
		base := strings.ToLower(args[0])
		list, _ := LoadBlockList()

		idx := slices.Index(list, base)
		if idx == -1 {
			isJSON, _ := cmd.Flags().GetBool("json")
			Reply(isJSON, "not found", base)
			return
		}

		// delete element
		list = slices.Delete(list, idx, idx+1)
		if err := SaveBlockList(list); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}

		isJSON, _ := cmd.Flags().GetBool("json")
		Reply(isJSON, "removed", base)
	},
}
