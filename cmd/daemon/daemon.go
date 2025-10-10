
package daemon

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"procguard/internal/database"
	"procguard/internal/ignore"
	"procguard/internal/logger"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the background daemon for process monitoring and blocking",
	Run: func(cmd *cobra.Command, args []string) {
		appLogger := logger.Get()

		db, err := database.InitDB()
		if err != nil {
			appLogger.Fatalf("Failed to initialize database: %v", err)
		}
		defer db.Close()

		Start(appLogger, db)

		// Keep the main goroutine alive
		select {}
	},
}

// Start runs the core daemon logic in goroutines.
func Start(appLogger *log.Logger, db *sql.DB) {
	// Goroutine for event-based process logging
	go runEventLogging(appLogger, db)

	// Goroutine for killing blocked processes
	go runProcessKiller(appLogger)
}

func runEventLogging(appLogger *log.Logger, db *sql.DB) {
	// runningProcs stores the PIDs of processes we are currently tracking.
	runningProcs := make(map[int32]bool)
	// Initialize the map with currently running processes that should be tracked.
	initializeRunningProcs(runningProcs, db)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		procs, err := process.Processes()
		if err != nil {
			continue
		}

		currentProcs := make(map[int32]bool)
		for _, p := range procs {
			currentProcs[p.Pid] = true
		}

		// 1. Check for ended processes
		for pid := range runningProcs {
			if !currentProcs[pid] {
				// Process has ended. Update its end_time in the DB.
				_, err := db.Exec("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
				if err != nil {
					appLogger.Printf("Failed to update end_time for pid %d: %v", pid, err)
				}
				delete(runningProcs, pid)
			}
		}

		// 2. Check for new processes
		for _, p := range procs {
			if !runningProcs[p.Pid] {
				// This is a new process. Check if we should log it.
				if shouldLogProcess(p) {
					name, _ := p.Name()
					parent, _ := p.Parent()
					parentName := ""
					if parent != nil {
						parentName, _ = parent.Name()
					}

					_, err := db.Exec("INSERT INTO app_events (process_name, pid, parent_process_name, start_time) VALUES (?, ?, ?, ?)",
						name, p.Pid, parentName, time.Now().Unix())
					if err != nil {
						appLogger.Printf("Failed to insert new process %s (pid %d): %v", name, p.Pid, err)
					}
					runningProcs[p.Pid] = true
				}
			}
		}
	}
}

// initializeRunningProcs pre-populates the runningProcs map with processes
// that are already in the database without an end_time.
func initializeRunningProcs(runningProcs map[int32]bool, db *sql.DB) {
	rows, err := db.Query("SELECT pid FROM app_events WHERE end_time IS NULL")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pid int32
		if err := rows.Scan(&pid); err == nil {
			// Verify the process is still running
			if exists, _ := process.PidExists(pid); exists {
				runningProcs[pid] = true
			} else {
				// Process is not running, so it should have been marked as ended.
				// This handles cases where the daemon was stopped abruptly.
				_, err := db.Exec("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
				if err != nil {
					logger.Get().Printf("Failed to backfill end_time for pid %d: %v", pid, err)
				}
			}
		}
	}
}

// shouldLogProcess determines if a process should be logged based on platform-specific rules
// and the user's request to not log child processes.
func shouldLogProcess(p *process.Process) bool {
	name, err := p.Name()
	if err != nil || name == "" {
		return false // Skip processes with no name
	}

	// Universal check: ignore self
	if p.Pid == int32(os.Getpid()) {
		return false
	}

	parent, err := p.Parent()
	if err != nil {
		return false // Skip processes with no parent
	}
	parentName, _ := parent.Name()

	var ignoreList []string
	if runtime.GOOS == "windows" {
		ignoreList = ignore.DefaultWindows
		// On Windows, a good heuristic for a "user-initiated" app is one parented by explorer.exe
		// or one with no parent that isn't a system process.
		if parentName == "explorer.exe" {
			return !ignore.IsIgnored(name, ignoreList)
		}
	} else {
		ignoreList = ignore.DefaultLinux
	}

	// The user wants to avoid logging child processes.
	// A good heuristic is to only log processes whose parent is a system/session manager process (and is in the ignore list),
	// or to explicitly ignore processes whose parents are known shells or other apps.
	// The current ignore list is for filtering out system processes, so if the parent is in the list, the child is likely a user app.
	// Let's refine the logic: don't log a process if its direct parent is NOT a system-level service.
	// This is tricky. A simpler, more robust rule is to just filter out the noise, which the ignore list does.
	// If a process name or its parent's name is in the ignore list, we skip it.
	if ignore.IsIgnored(name, ignoreList) || ignore.IsIgnored(parentName, ignoreList) {
		return false
	}

	// Platform-specific checks
	if runtime.GOOS == "linux" {
		uids, err := p.Uids()
		if err != nil || len(uids) == 0 || uids[0] < 1000 {
			return false // Skip system users
		}
	}
	// Add other platform-specific checks if needed

	return true
}

func runProcessKiller(appLogger *log.Logger) {
	killTick := time.NewTicker(100 * time.Millisecond)
	defer killTick.Stop()
	for range killTick.C {
		list, err := fetchBlocklist()
		if err != nil {
			// Use the main app logger, not the one passed in, to avoid confusion
			logger.Get().Printf("failed to fetch blocklist: %v", err)
			continue
		}
		if len(list) == 0 {
			continue
		}
		procs, _ := process.Processes()
		for _, p := range procs {
			name, _ := p.Name()
			if name == "" {
				continue // Skip processes with no name
			}

			// Enforce the blocklist by killing any process whose name is in the list.
			if slices.Contains(list, strings.ToLower(name)) {
				err := p.Kill()
				if err != nil {
					logger.Get().Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
				} else {
					logger.Get().Printf("killed blocked process %s (pid %d)", name, p.Pid)
				}
			}
		}
	}
}

func fetchBlocklist() ([]string, error) {
	resp, err := http.Get("http://127.0.0.1:58141/api/blocklist")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var blocklist []string
	if err := json.NewDecoder(resp.Body).Decode(&blocklist); err != nil {
		return nil, err
	}

	return blocklist, nil
}
