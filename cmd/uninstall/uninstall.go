package uninstall

import (
	"bufio"
	"fmt"
	"os"
	"procguard/cmd/auth"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
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

	fmt.Println("Stopping all ProcGuard services...")
	if err := killOtherProcGuardProcesses(); err != nil {
		// Log a warning but continue with uninstallation
		fmt.Fprintf(os.Stderr, "Warning: could not kill all running processes: %v\n", err)
	}

	fmt.Println("Uninstalling ProcGuard...")
	return platformUninstall()
}

func killOtherProcGuardProcesses() error {
	currentPid := os.Getpid()
	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("failed to get processes: %w", err)
	}

	for _, p := range procs {
		if p.Pid == int32(currentPid) {
			continue
		}

		name, err := p.Name()
		if err != nil {
			// Ignore processes we can't get the name of
			continue
		}

		if strings.HasPrefix(strings.ToLower(name), "procguard") {
			fmt.Printf("Killing running ProcGuard process: %s (pid %d)\n", name, p.Pid)
			if err := p.Kill(); err != nil {
				// Log a warning but continue
				fmt.Fprintf(os.Stderr, "Warning: failed to kill process %s (pid %d): %v\n", name, p.Pid, err)
			}
		}
	}
	return nil
}
