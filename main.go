package main

//go:generate go run github.com/akavel/rsrc -manifest build/procguard.manifest -o build/cache/rsrc.syso

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"procguard/api"
	"procguard/gui"
	"procguard/internal/daemon"
	"procguard/internal/data"
	"procguard/internal/web"
	"strings"
	"syscall"
	"time"
)

func main() {
	// When launched by Chrome, the first argument is the extension's origin.
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		db, err := data.OpenDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		data.NewLogger(db)
		web.Run()
		return
	}

	// If no command-line arguments are provided, this is a default run (e.g., double-click).
	if len(os.Args) == 1 {
		HandleDefaultStartup()
	} else if len(os.Args) > 1 {
		switch os.Args[1] {
		case "run-api":
			runApi()
		case "run-daemon":
			runDaemon()
		}
	}
}

func runApi() {
	const defaultPort = "58141"
	addr := "127.0.0.1:" + defaultPort
	fmt.Println("Starting API server on http://" + addr)
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	data.NewLogger(db)
	api.StartWebServer(addr, registerWebRoutes)
}

func registerWebRoutes(srv *api.Server, r *http.ServeMux) {
	// Create a sub-filesystem rooted at the "frontend" directory.
	subFS, err := fs.Sub(gui.FrontendFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem for frontend: %v", err)
	}

	// Create a file server for our static assets from the sub-filesystem.
	staticFS := http.FileServer(http.FS(subFS))

	// Serve the new static file directories.
	r.Handle("/dist/", staticFS)
	r.Handle("/src/", staticFS)

	// Keep the existing template rendering handlers.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gui.HandleIndex(&srv.Mu, srv.IsAuthenticated, srv.Logger, w, r)
	})
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		gui.HandleLoginTemplate(srv.Logger, w, r)
	})
	r.HandleFunc("/ping", gui.HandlePing)
}

func runDaemon() {
	fmt.Println("Starting ProcGuard daemon...")
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	data.NewLogger(db)
	appLogger := data.GetLogger()
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.Printf("Failed to close database: %v", err)
		}
	}()

	daemon.Start(appLogger, db)
	// Keep the main goroutine alive
	select {}
}

// HandleDefaultStartup implements the main startup logic for GUI mode on Windows.
func HandleDefaultStartup() {
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	data.NewLogger(db)

	// The autostart logic has been moved to a user-initiated action in the GUI.

	// Set up the native messaging host using the current executable's path.
	// Note: This means the original executable must be kept.
	// This will be improved when a proper installer is built.
	exePath, err := os.Executable()
	if err != nil {
		data.GetLogger().Printf("Error getting executable path: %v", err)
	}
	if err := web.InstallNativeHost(exePath); err != nil {
		data.GetLogger().Printf("Failed to install native messaging host: %v\n", err)
		// We don't want to block the main application from starting if this fails.
	}

	const defaultPort = "58141"
	guiAddress := "127.0.0.1:" + defaultPort
	guiUrl := "http://" + guiAddress

	// Check if a server is already running
	_, err = http.Get(guiUrl + "/ping")
	if err == nil {
		// Instance is already running. Just open the browser and exit.
		openBrowser(guiUrl)
		return
	}

	// No instance is running. This is the first instance.
	fmt.Println("Starting ProcGuard services...")

	// Start the API and daemon services in the background.
	cmdApi := exec.Command(exePath, "run-api")
	cmdApi.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	if err := cmdApi.Start(); err != nil {
		data.GetLogger().Printf("Error starting API service: %v", err)
		return
	}

	cmdDaemon := exec.Command(exePath, "run-daemon")
	cmdDaemon.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	if err := cmdDaemon.Start(); err != nil {
		data.GetLogger().Printf("Error starting daemon service: %v", err)
		return
	}

	// Give the server a moment to start before opening the browser.
	time.Sleep(1 * time.Second)
	openBrowser(guiUrl)
}

func openBrowser(url string) {
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		data.GetLogger().Printf("Error opening browser: %v\n", err)
	}
}
