package cmd

import (
	"procguard/cmd/block"
	"procguard/cmd/daemon"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "procguard",
	Short: "Process monitor and control program",
}

func Execute() { cobra.CheckErr(rootCmd.Execute()) }

func init() {
	rootCmd.AddCommand(daemon.DaemonCmd)
	rootCmd.AddCommand(block.BlockCmd)
}
