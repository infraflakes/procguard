//go:build windows

package cmd

import (
	"os/exec"
)

func openBrowser(url string) {
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
