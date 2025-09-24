package cmd

import (
	"fmt"
	"net/http"
	"procguard/cmd/block"
	"procguard/cmd/daemon"
	"procguard/cmd/gui"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "procguard",
	Short: "Process monitor and control program",
	Run: func(cmd *cobra.Command, args []string) {
		// This runs when no subcommand is given
		const defaultPort = "58141"
		guiAddress := "127.0.0.1:" + defaultPort
		guiUrl := "http://" + guiAddress

		// Check if a server is already running
		_, err := http.Get(guiUrl + "/ping")
		if err == nil {
			// Instance is already running. Just open the browser and exit.
			fmt.Println("ProcGuard is already running. Opening GUI...")
			openBrowser(guiUrl)
			return
		}

		// No instance is running. This is the first instance.
		fmt.Println("Starting ProcGuard...")

		// Set up autostart for Windows if applicable.
		daemon.EnsureAutostartTask()

		// Start the daemon in the background
		go daemon.Start()

		// Open the browser in a goroutine so it doesn't block.
		go func() {
			// Add a small delay to give the server time to start listening.
			time.Sleep(1 * time.Second)
			openBrowser(guiUrl)
		}()

		// Start the web server (this is a blocking call)
		gui.StartWebServer(guiAddress)
	},
}

func Execute() { cobra.CheckErr(rootCmd.Execute()) }

func init() {
	rootCmd.AddCommand(daemon.DaemonCmd)
	rootCmd.AddCommand(block.BlockCmd)
	rootCmd.AddCommand(gui.GuiCmd)
}
