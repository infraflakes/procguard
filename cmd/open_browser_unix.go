//go:build linux || darwin

package cmd

import (
	"log"
	"os/exec"
)

func openBrowser(url string) {
	if err := exec.Command("xdg-open", url).Start(); err != nil {
		log.Printf("Failed to open browser for %s: %v", url, err)
	}
}
