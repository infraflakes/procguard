package app

import (
	"database/sql"
	"procguard/internal/data"
	"slices"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	processCheckInterval     = 2 * time.Second
	blocklistEnforceInterval = 2 * time.Second
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")

	enumWindowsCallback = syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		params := (*enumWindowsParams)(unsafe.Pointer(lParam))
		var windowPid uint32
		procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&windowPid)))

		if windowPid == params.pid {
			if isVisible, _, _ := procIsWindowVisible.Call(uintptr(hwnd)); isVisible != 0 {
				params.found = true
				return 0 // Stop enumeration
			}
		}
		return 1 // Continue
	})
)

type enumWindowsParams struct {
	pid   uint32
	found bool
}

// hasVisibleWindow checks if a process with the given PID has a visible window.
func hasVisibleWindow(pid uint32) bool {
	params := &enumWindowsParams{pid: pid, found: false}
	procEnumWindows.Call(enumWindowsCallback, uintptr(unsafe.Pointer(params)))
	return params.found
}

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
			data.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
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
				data.EnqueueWrite("INSERT INTO app_events (process_name, pid, parent_process_name, exe_path, start_time) VALUES (?, ?, ?, ?, ?)",
					name, p.Pid, parentName, exePath, time.Now().Unix())
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
				data.EnqueueWrite("UPDATE app_events SET end_time = ? WHERE pid = ? AND end_time IS NULL", time.Now().Unix(), pid)
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

	// Do not log the ProcGuard process itself or other ignored processes.
	if IsIgnored(name, DefaultWindows) || IsIgnored(name, []string{"ProcGuardSvc.exe"}) {
		return false
	}

	// Log if it has a visible window.
	if hasVisibleWindow(uint32(p.Pid)) {
		return true
	}

	// If it doesn't have a window, check its integrity level.
	il, err := GetProcessIntegrityLevel(uint32(p.Pid))
	if err == nil && il >= SECURITY_MANDATORY_SYSTEM_RID {
		return false // Skip system and high integrity processes.
	}

	parent, err := p.Parent()
	if err != nil {
		// No parent and no window, could be a standalone background task. Log it.
		return true
	}

	parentName, err := parent.Name()
	if err != nil {
		return true // Can't get parent name, assume it's a top-level process.
	}

	// If the parent is a known system process, don't log the child.
	if IsIgnored(parentName, DefaultWindows) {
		return false
	}

	// By default, do not log child processes without a visible window.
	return false
}
