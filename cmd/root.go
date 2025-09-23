package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "procguard",
	Short: "Process monitor and control program",
}

func Execute() { cobra.CheckErr(rootCmd.Execute()) }
