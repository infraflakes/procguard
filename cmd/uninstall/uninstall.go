package uninstall

import (
	"bufio"
	"fmt"
	"os"
	"procguard/cmd/auth"
	"strings"

	"github.com/spf13/cobra"
)

var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall ProcGuard from the system",
	RunE:  runUninstall,
}

func init() {
	UninstallCmd.Flags().Bool("force-no-prompt", false, "Force uninstall without prompt")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force-no-prompt")

	if !force {
		if err := auth.Check(); err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Are you sure you want to completely uninstall ProcGuard? This action cannot be undone. [y/N]: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))

		if text != "y" && text != "yes" {
			fmt.Println("Uninstall cancelled.")
			return nil
		}
	}

	fmt.Println("Uninstalling ProcGuard...")
	return platformUninstall()
}
