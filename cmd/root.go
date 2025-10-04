package cmd

import (
	"procguard/cmd/block"
	"procguard/cmd/run"
	"procguard/cmd/uninstall"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "procguard",
	Short: "Process monitor and control program",
}

func Execute() { cobra.CheckErr(rootCmd.Execute()) }

func init() {
	rootCmd.AddCommand(run.RunCmd)
	rootCmd.AddCommand(block.BlockCmd)
	rootCmd.AddCommand(uninstall.UninstallCmd)
}
