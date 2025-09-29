//go:build linux

package daemon

import (
	"log"
	"procguard/internal/ignore"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func runLogging(logger *log.Logger) {
	ignoreList := ignore.DefaultLinux
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

			uids, err := p.Uids()
			if err != nil || len(uids) == 0 {
				continue // Skip if we can't get UID
			}
			// Stage 1: Skip system users (UID < 1000)
			if uids[0] < 1000 {
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

			// Stage 2: Skip user-level system processes by name
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
