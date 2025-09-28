package block

import (
	"encoding/json"
	"fmt"
	"os"
	"procguard/internal/blocklist"
	"slices"

	"github.com/spf13/cobra"
)

// BlockLoadCmd defines the command for loading a blocklist from a file.
var BlockLoadCmd = &cobra.Command{
	Use:   "load <file>",
	Short: "Load block-list from a file, merging with the existing list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Read the content of the file to be loaded.
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("load: %w", err)
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
				return fmt.Errorf("load: invalid JSON format in %s", filePath)
			}
			newEntries = savedList.Blocked
		}

		// Load the existing blocklist to merge with.
		existingList, err := blocklist.Load()
		if err != nil {
			return err
		}

		// Merge the two lists, ensuring that there are no duplicate entries.
		for _, entry := range newEntries {
			if !slices.Contains(existingList, entry) {
				existingList = append(existingList, entry)
			}
		}

		// Save the newly merged list.
		if err := blocklist.Save(existingList); err != nil {
			return fmt.Errorf("save: %w", err)
		}

		fmt.Println("block-list loaded and merged from:", filePath)
		return nil
	},
}
