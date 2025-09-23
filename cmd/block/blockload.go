package block

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

// BlockLoadCmd defines the command for loading a blocklist from a file.
var BlockLoadCmd = &cobra.Command{
	Use:   "load <file>",
	Short: "Load block-list from a file, merging with the existing list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// Read the content of the file to be loaded.
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "load:", err)
			os.Exit(1)
		}

		// The file can be either a simple JSON array of strings or the format
		// produced by the 'block save' command. We need to handle both cases.
		var newEntries []string
		var savedList struct {
			Blocked []string `json:"blocked"`
		}

		// First, try to unmarshal as a simple list of strings.
		err = json.Unmarshal(content, &newEntries)
		if err != nil {
			// If that fails, try to unmarshal as the 'block save' format.
			err2 := json.Unmarshal(content, &savedList)
			if err2 != nil {
				// If both fail, the file is not a valid blocklist.
				fmt.Fprintln(os.Stderr, "load: invalid JSON format in", filePath)
				os.Exit(1)
			}
			newEntries = savedList.Blocked
		}

		// Load the existing blocklist to merge with.
		existingList, _ := LoadBlockList()

		// Merge the two lists, ensuring that there are no duplicate entries.
		for _, entry := range newEntries {
			if !slices.Contains(existingList, entry) {
				existingList = append(existingList, entry)
			}
		}

		// Save the newly merged list.
		if err := SaveBlockList(existingList); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}

		fmt.Println("block-list loaded and merged from:", filePath)
	},
}
