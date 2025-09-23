package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const tick = 3 * time.Second

func main() {
	// find os cache dir (~/.cache on linux, %LOCALAPPDATA% on win)
	cacheDir, _ := os.UserCacheDir()
	logFile := filepath.Join(cacheDir, "procguard", "events.log")

	// make sure folder exists or next line will cry
	os.MkdirAll(filepath.Dir(logFile), 0755)

	// open log : create if missing, append if exists, 0644 so others can read
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// hijack log pkg so we can use log.Printf without timestamp prefix
	log.SetOutput(f)
	log.SetFlags(0)

	// ticker fires every 3s; range blocks forever
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for range ticker.C {
		// fetch all running processes (cheap call, no root needed)
		procs, _ := process.Processes()
		for _, p := range procs {
			name, _ := p.Name()
			if name == "" {
				continue // skip ghosts
			}
			// line format : unixtime|exe|pid  (easy to grep later)
			log.Printf("%d|%s|%d\n", time.Now().Unix(), name, p.Pid)
		}
	}
}
