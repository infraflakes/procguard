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

func runDaemon(cmd *cobra.Command, args []string) {
	// cache dir: ~/.cache on linux, %LOCALAPPDATA% on win
	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	// make sure folder exists so openFile doesn't cry
	os.MkdirAll(filepath.Dir(logFile), 0755)

	// open log : create | append | 0644 let others read
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// readable format logger
	logger := log.New(f, "", 0)

	// 3s ticker; range blocks forever
	tick := time.NewTicker(3 * time.Second)
	defer tick.Stop()

	for range tick.C {
		// all pids, no root needed for obvious reasons
		procs, _ := process.Processes()
		list, _ := block.LoadBlockList()

		for _, p := range procs {
			name, _ := p.Name()
			if name == "" {
				continue // skip ghosts
			}

			parent, _ := p.Parent()
			parentName, _ := parent.Name()

			// pretty time eg. 2025-09-24 19:47:05 | exe | pid | parent_exe
			logger.Printf("%s | %s | %d | %s\n",
				time.Now().Format("2006-01-02 15:04:05"),
				name,
				p.Pid,
				parentName)

			// kill if blocked
			if slices.Contains(list, strings.ToLower(name)) {
				err := p.Kill()
				if err != nil {
					logger.Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
				} else {
					logger.Printf("killed blocked process %s (pid %d)", name, p.Pid)
				}
			}
		}
	}
}
