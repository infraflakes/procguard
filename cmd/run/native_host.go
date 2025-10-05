package run

import (
	"fmt"
	"procguard/internal/native"

	"github.com/spf13/cobra"
)

var nativeHostCmd = &cobra.Command{
	Use:   "native-host",
	Short: "Run the ProcGuard native messaging host",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting ProcGuard native messaging host...")
		native.Run()
	},
}
