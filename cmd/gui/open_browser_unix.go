//go:build linux || darwin

package gui

import (
	"os/exec"
)

func openBrowser(url string) {
	exec.Command("xdg-open", url).Start()
}
