package block

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// BlockSaveCmd defines the command for saving the current blocklist to a file.
var BlockSaveCmd = &cobra.Command{
	Use:   "save <file>",
	Short: "Save current block-list to chosen path (CLI: specify path; GUI: will hook native dialog)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckAuth(cmd)
		list, _ := LoadBlockList()
		dest := args[0]

		// The blocklist is saved in a JSON object that includes a timestamp for when
		// the export was created. This provides some context for the exported data.
		header := map[string]interface{}{
			"exported_at": time.Now().Format(time.RFC3339),
			"blocked":     list,
		}

		// Marshal the data to JSON with indentation for readability.
		b, _ := json.MarshalIndent(header, "", "  ")
		if err := os.WriteFile(dest, b, 0644); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}

		fmt.Println("saved to:", dest)
	},
}
