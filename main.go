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

	// keep our own logger so we can keep timestamp ON for pretty date
	logger := log.New(f, "", 0) // no extra prefix

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
			// ---- parent hunt ----
			parentName := "<unknown>"
			if pp, err := p.Parent(); err == nil && pp != nil {
				parentName, _ = pp.Name()
			}
			// pretty line: 2025-09-24 19:47:05 | Parent - Child | pid
			logger.Printf("%s | %s - %s | %d\n",
				time.Now().Format("2006-01-02 15:04:05"),
				parentName,
				name,
				p.Pid)
		}
	}
}
