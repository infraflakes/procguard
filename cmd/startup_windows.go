//go:build windows

package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"procguard/cmd/daemon"
	"syscall"
	"time"
)

// HandleDefaultStartup implements the main startup logic for GUI mode on Windows.
func HandleDefaultStartup() {
	const defaultPort = "58141"
	guiAddress := "127.0.0.1:" + defaultPort
	guiUrl := "http://" + guiAddress

	// Check if a server is already running
	_, err := http.Get(guiUrl + "/ping")
	if err == nil {
		// Instance is already running. Just open the browser and exit.
	openBrowser(guiUrl)
		return
	}

	// No instance is running. This is the first instance.
	fmt.Println("Starting ProcGuard services...")

	// Set up autostart for Windows if applicable.

daemon.EnsureAutostartTask()

	// Start the API and daemon services in the background.
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
		return
	}

	cmdApi := exec.Command(exePath, "run", "api")
	cmdApi.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	if err := cmdApi.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error starting API service:", err)
		return
	}

	cmdDaemon := exec.Command(exePath, "run", "daemon")
	cmdDaemon.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	if err := cmdDaemon.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Error starting daemon service:", err)
		return
	}

	// Give the server a moment to start before opening the browser.
	time.Sleep(1 * time.Second)
	openBrowser(guiUrl)
}

func openBrowser(url string) {
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening browser: %v\n", err)
	}
}
