//go:build windows

package cmd

import (
	"os/exec"
	"procguard/internal/config"
)

func checkAutostart() {
	cfg, err := config.Load()
	if err != nil {
		return // Can't load config, so can't check autostart
	}

	if cfg.AutostartEnabled {
		// Check if the task exists
		err := exec.Command("schtasks", "/query", "/tn", taskName).Run()
		if err != nil { // Task likely doesn't exist
			installAutostartTask(nil, nil)
		}
	}
}
