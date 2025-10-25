package main

//go:generate go run github.com/akavel/rsrc -manifest build/procguard.manifest -o build/cache/rsrc.syso

import (
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

const (
	// defaultPort is the port used by the web server.
	defaultPort = "58141"
	// chromeExtensionID is the ID of the Chrome extension that communicates with the native messaging host.
	chromeExtensionID = "ilaocldmkhlifnikhinkmiepekpbefoh"

	// command line arguments to run specific parts of the application.
	cmdRunAPI    = "run-api"
	cmdRunDaemon = "run-daemon"
)

// main is the entry point of the application. It determines the execution mode based on command-line arguments.
func main() {
	// When the application is launched by Chrome as a native messaging host,
	// the first argument is the origin of the extension.
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		runNativeMessagingHost()
		return
	}

	// If there are no arguments, start the main GUI application.
	// Otherwise, run the specified command.
	if len(os.Args) == 1 {
		startGUIApplication()
	} else {
		switch os.Args[1] {
		case cmdRunAPI:
			startAPIServer()
		case cmdRunDaemon:
			startDaemonService()
		}
	}
}

// runNativeMessagingHost starts the application in native messaging host mode,
// allowing communication with the browser extension.
func runNativeMessagingHost() {
	db, err := data.OpenDB()
	if err != nil {
		log.Fatalf("Native messaging host failed to open database: %v", err)
	}
	data.NewLogger(db)
	web.Run()
}

// startAPIServer initializes and starts the API server.
func startAPIServer() {
	addr := "127.0.0.1:" + defaultPort

	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database for API server: %v", err)
	}
	data.NewLogger(db)
	api.StartWebServer(addr, registerWebRoutes)
}

// registerWebRoutes sets up the routes for the web server.
func registerWebRoutes(srv *api.Server, r *http.ServeMux) {
	// Create a sub-filesystem for the frontend assets.
	subFS, err := fs.Sub(gui.FrontendFS, "frontend")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem for frontend: %v", err)
	}

	staticFS := http.FileServer(http.FS(subFS))

	// Serve static assets.
	r.Handle("/dist/", staticFS)
	r.Handle("/src/", staticFS)

	// Serve application pages.
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gui.HandleIndex(&srv.Mu, srv.IsAuthenticated, srv.Logger, w, r)
	})
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		gui.HandleLoginTemplate(srv.Logger, w, r)
	})
	r.HandleFunc("/ping", gui.HandlePing)
}

// startDaemonService initializes and starts the background daemon.
func startDaemonService() {
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database for daemon: %v", err)
	}
	data.NewLogger(db)
	defer func() {
		if err := db.Close(); err != nil {
			data.GetLogger().Printf("Failed to close database in daemon: %v", err)
		}
	}()

	daemon.StartDaemon(data.GetLogger(), db)
	// Keep the daemon running indefinitely.
	select {}
}

// startGUIApplication handles the main startup logic for the GUI application.
func startGUIApplication() {
	db, err := data.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database for GUI: %v", err)
	}
	data.NewLogger(db)

	exePath, err := os.Executable()
	if err != nil {
		data.GetLogger().Printf("Error getting executable path: %v", err)
		// We can continue, but some features might not work.
	}

	// This setup is necessary for the browser extension to communicate with the application.
	if err := web.InstallNativeHost(exePath, chromeExtensionID); err != nil {
		data.GetLogger().Printf("Failed to install native messaging host: %v\n", err)
		// This is not a fatal error, the application can still run without the extension.
	}

	guiAddress := "127.0.0.1:" + defaultPort
	guiUrl := "http://" + guiAddress

	// Check if an instance of the application is already running.
	if isAppRunning(guiUrl) {
		openBrowser(guiUrl)
		return
	}

	// If no instance is running, start the API and daemon services as background processes.
	if err := startBackgroundService(exePath, cmdRunAPI); err != nil {
		data.GetLogger().Printf("Error starting API service: %v", err)
		return
	}
	if err := startBackgroundService(exePath, cmdRunDaemon); err != nil {
		data.GetLogger().Printf("Error starting daemon service: %v", err)
		return
	}

	// Give the server a moment to start before opening the browser.
	time.Sleep(1 * time.Second)
	openBrowser(guiUrl)
}

// isAppRunning checks if another instance of the application is already running by pinging the server.
func isAppRunning(url string) bool {
	resp, err := http.Get(url + "/ping")
	if err != nil {
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			data.GetLogger().Printf("Failed to close response body in isAppRunning: %v", err)
		}
	}()
	return resp.StatusCode == http.StatusOK
}

// startBackgroundService starts a command as a detached background process.
func startBackgroundService(path, command string) error {
	cmd := exec.Command(path, command)
	// CREATE_NO_WINDOW flag prevents a console window from appearing on Windows.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	return cmd.Start()
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) {
	if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start(); err != nil {
		data.GetLogger().Printf("Error opening browser: %v\n", err)
	}
}
