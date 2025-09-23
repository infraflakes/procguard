package cmd

import (
	"log"
	"os"
	"path/filepath"
	"procguard/cmd/block"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

func init() { rootCmd.AddCommand(daemonCmd) }

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run in the background, logs every 3 seconds",
	Run:   runDaemon,
}

// runDaemon starts the background process that monitors and logs system processes.
// It periodically fetches the list of running processes, logs them, and terminates
// any process found in the blocklist.
func runDaemon(cmd *cobra.Command, args []string) {
	checkAutostart()

	// Determine the appropriate cache directory based on the user's OS.
	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	// Ensure the directory for the log file exists before trying to create the file.
	os.MkdirAll(filepath.Dir(logFile), 0755)

	// Open the log file for appending, creating it if it doesn't exist.
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Create a logger that writes to the log file in a simple, readable format.
	logger := log.New(f, "", 0)

	// Use a ticker to trigger the process scan at regular intervals.
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()

	// The main loop of the daemon, which runs indefinitely.
	for range tick.C {
		// Get a list of all running processes on the system.
		procs, _ := process.Processes()
		// Load the blocklist to check against running processes.
		list, _ := block.LoadBlockList()

		for _, p := range procs {
			name, _ := p.Name()
			if name == "" {
				continue // Skip processes with no name (e.g., kernel processes).
			}

			// Get the parent process information for more detailed logging.
			parent, _ := p.Parent()
			parentName, _ := parent.Name()

			// Log the process information in a structured format.
			logger.Printf("%s | %s | %d | %s\n",
				time.Now().Format("2006-01-02 15:04:05"),
				name,
				p.Pid,
				parentName)

			// Enforce the blocklist by killing any process whose name is in the list.
			if slices.Contains(list, strings.ToLower(name)) {
				err := p.Kill()
				if err != nil {
					// Log any errors that occur during process termination.
					logger.Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
				} else {
					logger.Printf("killed blocked process %s (pid %d)", name, p.Pid)
				}
			}
		}
	}
}
