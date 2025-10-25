package app

import (
	"database/sql"
	"os"
	"path/filepath"
	"procguard/internal/data"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	processCheckInterval     = 2 * time.Second
	blocklistEnforceInterval = 100 * time.Millisecond
)

// StartProcessEventLogger starts a long-running goroutine that monitors process creation and termination events.
func StartProcessEventLogger(appLogger data.Logger, db *sql.DB) {
	go func() {
		// runningProcs stores the PIDs of processes we are currently tracking.
		runningProcs := make(map[int32]bool)
		// Initialize the map with currently running processes that should be tracked.
		initializeRunningProcs(runningProcs, db)

		ticker := time.NewTicker(processCheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			procs, err := process.Processes()
			if err != nil {
				appLogger.Printf("Failed to get processes: %v", err)
				continue
			}

			currentProcs := make(map[int32]bool)
			for _, p := range procs {
				currentProcs[p.Pid] = true
			}

			logEndedProcesses(appLogger, db, runningProcs, currentProcs)
			logNewProcesses(appLogger, db, runningProcs, procs)
		}
	}()
}

// logEndedProcesses checks for processes that have terminated and updates their end time in the database.
func logEndedProcesses(appLogger data.Logger, db *sql.DB, runningProcs, currentProcs map[int32]bool) {
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
}

// logNewProcesses checks for new processes and logs them to the database if they should be tracked.
func logNewProcesses(appLogger data.Logger, db *sql.DB, runningProcs map[int32]bool, procs []*process.Process) {
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

				exePath, err := p.Exe()
				if err != nil {
					appLogger.Printf("Failed to get exe path for %s (pid %d): %v", name, p.Pid, err)
				}
				_, err = db.Exec("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time) VALUES (?, ?, ?, ?, ?)",
					name, p.Pid, parentName, exePath, time.Now().Unix())
				if err != nil {
					appLogger.Printf("Failed to insert new process %s (pid %d): %v", name, p.Pid, err)
				}
				runningProcs[p.Pid] = true
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
	defer func() {
		if err := rows.Close(); err != nil {
			data.GetLogger().Printf("Failed to close rows: %v", err)
		}
	}()

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
					data.GetLogger().Printf("Failed to backfill end_time for pid %d: %v", pid, err)
				}
			}
		}
	}
}

// StartBlocklistEnforcer starts a long-running goroutine that periodically checks for and kills blocked processes.
func StartBlocklistEnforcer(appLogger data.Logger) {
	go func() {
		killTick := time.NewTicker(blocklistEnforceInterval)
		defer killTick.Stop()
		for range killTick.C {
			list, err := data.LoadAppBlocklist()
			if err != nil {
				appLogger.Printf("failed to fetch blocklist: %v", err)
				continue
			}
			if len(list) == 0 {
				continue
			}
			procs, err := process.Processes()
			if err != nil {
				appLogger.Printf("Failed to get processes: %v", err)
				continue
			}
			for _, p := range procs {
				name, _ := p.Name()
				if name == "" {
					continue // Skip processes with no name
				}

				// Enforce the blocklist by killing any process whose name is in the list.
				if slices.Contains(list, strings.ToLower(name)) {
					err := p.Kill()
					if err != nil {
						appLogger.Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
					} else {
						appLogger.Printf("killed blocked process %s (pid %d)", name, p.Pid)
					}
				}
			}
		}
	}()
}

// shouldLogProcess determines if a process should be logged based on a set of heuristics
// designed to filter out system and other irrelevant processes.
func shouldLogProcess(p *process.Process) bool {
	name, err := p.Name()
	if err != nil || name == "" {
		return false // Skip processes with no name
	}

	// Do not log the ProcGuard process itself.
	if p.Pid == int32(os.Getpid()) {
		return false
	}

	parent, err := p.Parent()
	if err != nil {
		return false // Skip processes with no parent
	}
	parentName, _ := parent.Name()

	// Do not log a process if its parent has the same name, as these are often helper processes.
	if name == parentName {
		return false
	}

	// Do not log a process if it's in the same directory as its parent, as these are also often helper processes.
	childExe, err := p.Exe()
	if err == nil {
		parentExe, err := parent.Exe()
		if err == nil {
			if filepath.Dir(childExe) == filepath.Dir(parentExe) {
				return false
			}
		}
	}

	// On Windows, filter out processes based on their integrity level and a predefined ignore list.
	il, err := GetProcessIntegrityLevel(uint32(p.Pid))
	if err == nil && il >= SECURITY_MANDATORY_SYSTEM_RID {
		return false // Skip system and high integrity processes.
	}
	if IsIgnored(name, DefaultWindows) || IsIgnored(parentName, DefaultWindows) {
		return false
	}

	return true
}
