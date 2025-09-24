package daemon

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

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run in the background, logs every 15 seconds",
	Run: func(cmd *cobra.Command, args []string) {
		Start()
		// Keep the main goroutine alive
		select {}
	},
}

// Start runs the core daemon logic in goroutines.
func Start() {

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
	// This defer will likely not be called if the daemon runs indefinitely, but it's good practice.
	// In a real-world scenario, you'd have a proper shutdown mechanism.
	// defer f.Close()

	// Create a logger that writes to the log file in a simple, readable format.
	logger := log.New(f, "", 0)

	// Goroutine for logging processes
	go func() {
		logTick := time.NewTicker(15 * time.Second)
		defer logTick.Stop()
		for range logTick.C {
			seen := make(map[int32]bool)
			procs, _ := process.Processes()
			for _, p := range procs {
				if seen[p.Pid] {
					continue
				}
				seen[p.Pid] = true

				name, _ := p.Name()
				if name == "" || strings.HasPrefix(name, "kworker") || name == "kthreadd" {
					continue // Skip processes with no name or kworker threads.
				}

				// Get the parent process information for more detailed logging.
				parent, err := p.Parent()
				if err != nil {
					continue // Skip processes with no parent
				}
				parentName, _ := parent.Name()

				// Log the process information in a structured format.
				logger.Printf("%s | %s | %d | %s\n",
					time.Now().Format("2006-01-02 15:04:05"),
					name,
					p.Pid,
					parentName)
			}
		}
	}()

	// Goroutine for killing blocked processes
	go func() {
		killTick := time.NewTicker(100 * time.Millisecond)
		defer killTick.Stop()
		for range killTick.C {
			list, _ := block.LoadBlockList()
			if len(list) == 0 {
				continue
			}
			procs, _ := process.Processes()
			for _, p := range procs {
				name, _ := p.Name()
				if name == "" {
					continue // Skip processes with no name (e.g., kernel processes).
				}

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
	}()
}
