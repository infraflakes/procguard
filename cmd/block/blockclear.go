package block

import (
	"fmt"
	"procguard/internal/client"

	"github.com/spf13/cobra"
)

var BlockClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Empty the block-list completely",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New()
		if err := c.ClearBlocklist(); err != nil {
			return err
		}
		fmt.Println("block-list cleared")
		return nil
	},
}
