//go:build windows

package gui

import (
	"os/exec"
)

func openBrowser(url string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
