//go:build windows

package daemon

import (
	"log"
	"procguard/internal/ignore"
	"procguard/internal/winutil"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func runLogging(logger *log.Logger) {
	ignoreList := ignore.DefaultWindows
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

			// Stage 1: Skip processes with System or High integrity level
			il, err := winutil.GetProcessIntegrityLevel(uint32(p.Pid))
			if err == nil && il >= winutil.SECURITY_MANDATORY_SYSTEM_RID {
				continue
			}

			name, _ := p.Name()
			if name == "" {
				continue // Skip processes with no name
			}

			parent, err := p.Parent()
			if err != nil {
				continue // Skip processes with no parent
			}
			parentName, _ := parent.Name()

			// Stage 2: Skip based on name
			if ignore.IsIgnored(name, ignoreList) || ignore.IsIgnored(parentName, ignoreList) {
				continue
			}

			logger.Printf("%s | %s | %d | %s\n",
				time.Now().Format("2006-01-02 15:04:05"),
				name,
				p.Pid,
				parentName)
		}
	}
}
