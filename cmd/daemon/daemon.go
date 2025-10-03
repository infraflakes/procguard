package daemon

import (
	"encoding/json"
	"log"
	"net/http"
	"procguard/internal/logger"
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
	logger := logger.Get()
	go runLogging(logger)

	// Goroutine for killing blocked processes
	go func(logger *log.Logger) {
		killTick := time.NewTicker(100 * time.Millisecond)
		defer killTick.Stop()
		for range killTick.C {
			list, err := fetchBlocklist()
			if err != nil {
				logger.Printf("failed to fetch blocklist: %v", err)
				continue
			}
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
	}(logger)
}

func fetchBlocklist() ([]string, error) {
	resp, err := http.Get("http://127.0.0.1:58141/api/blocklist")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Get().Printf("Error closing response body: %v", err)
		}
	}()

	var blocklist []string
	if err := json.NewDecoder(resp.Body).Decode(&blocklist); err != nil {
		return nil, err
	}

	return blocklist, nil
}
