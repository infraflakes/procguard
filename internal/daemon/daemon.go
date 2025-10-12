package daemon

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"procguard/internal/blocklist"
	"slices"
	"strings"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows"
)

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
					Get().Printf("Failed to backfill end_time for pid %d: %v", pid, err)
				}
			}
		}
	}
}

func runProcessKiller(appLogger *log.Logger) {
	killTick := time.NewTicker(100 * time.Millisecond)
	defer killTick.Stop()
	for range killTick.C {
		list, err := blocklist.LoadApp()
		if err != nil {
			// Use the main app logger, not the one passed in, to avoid confusion
			Get().Printf("failed to fetch blocklist: %v", err)
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
										Get().Printf("failed to kill %s (pid %d): %v", name, p.Pid, err)
									} else {
										Get().Printf("killed blocked process %s (pid %d)", name, p.Pid)
									}			}
		}
	}
}

// shouldLogProcess determines if a process should be logged based on platform-specific rules.
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

	// Do not log a process if its parent has the same name.
	if name == parentName {
		return false
	}

	// Windows-specific checks
	il, err := GetProcessIntegrityLevel(uint32(p.Pid))
	if err == nil && il >= SECURITY_MANDATORY_SYSTEM_RID {
		return false // Skip system/high integrity processes
	}
	if IsIgnored(name, DefaultWindows) || IsIgnored(parentName, DefaultWindows) {
		return false
	}

	return true
}

const (
	// Integrity Level constants
	SECURITY_MANDATORY_UNTRUSTED_RID         = 0x00000000
	SECURITY_MANDATORY_LOW_RID               = 0x00001000
	SECURITY_MANDATORY_MEDIUM_RID            = 0x00002000
	SECURITY_MANDATORY_HIGH_RID              = 0x00003000
	SECURITY_MANDATORY_SYSTEM_RID            = 0x00004000
	SECURITY_MANDATORY_PROTECTED_PROCESS_RID = 0x00005000
)

// GetProcessIntegrityLevel returns the integrity level of a process.
func GetProcessIntegrityLevel(pid uint32) (uint32, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pid)
	if err != nil {
		// Ignore errors for processes we can't open (e.g., system processes)
		return 0, nil
	}
	defer windows.Close(h)

	var token windows.Token
	if err := windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token); err != nil {
		return 0, fmt.Errorf("could not open process token: %w", err)
	}

	var tokenInfoLen uint32
	// First call to get the required buffer size. This is expected to fail.
	windows.GetTokenInformation(token, windows.TokenIntegrityLevel, nil, 0, &tokenInfoLen)
	if tokenInfoLen == 0 {
		return 0, fmt.Errorf("GetTokenInformation failed to get buffer size")
	}

	tokenInfo := make([]byte, tokenInfoLen)
	if err := windows.GetTokenInformation(token, windows.TokenIntegrityLevel, &tokenInfo[0], tokenInfoLen, &tokenInfoLen); err != nil {
		return 0, fmt.Errorf("could not get token information: %w", err)
	}

	til := (*windows.Tokenmandatorylabel)(unsafe.Pointer(&tokenInfo[0]))
	sid := til.Label.Sid

	// The integrity level is the last sub-authority in the SID.
	// A SID is structured as: [Revision][SubAuthorityCount][Authority][SubAuthority1]...[SubAuthorityN]
	// We need to get the address of the last SubAuthority.
	subAuthorityCount := *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(sid)) + 1))
	pSubAuthority := uintptr(unsafe.Pointer(sid)) + 8 + (uintptr(subAuthorityCount)-1)*4

	return *(*uint32)(unsafe.Pointer(pSubAuthority)), nil
}