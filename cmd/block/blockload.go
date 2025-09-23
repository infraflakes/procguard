package block

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

var BlockLoadCmd = &cobra.Command{
	Use:   "load <file>",
	Short: "Load block-list from a file, merging with the existing list",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// Read the source file
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "load:", err)
			os.Exit(1)
		}

		// Unmarshal the JSON. It can be a plain list or the format from 'block save'.
		var newEntries []string
		var savedList struct {
			Blocked []string `json:"blocked"`
		}

		// Try parsing as a plain list first
		err = json.Unmarshal(content, &newEntries)
		if err != nil {
			// If that fails, try parsing the 'block save' format
			err2 := json.Unmarshal(content, &savedList)
			if err2 != nil {
				fmt.Fprintln(os.Stderr, "load: invalid JSON format in", filePath)
				os.Exit(1)
			}
			newEntries = savedList.Blocked
		}

		// Load the existing list
		existingList, _ := LoadBlockList()

		// Merge lists and remove duplicates
		for _, entry := range newEntries {
			if !slices.Contains(existingList, entry) {
				existingList = append(existingList, entry)
			}
		}

		// Save the merged list
		if err := SaveBlockList(existingList); err != nil {
			fmt.Fprintln(os.Stderr, "save:", err)
			os.Exit(1)
		}

		fmt.Println("block-list loaded and merged from:", filePath)
	},
}
