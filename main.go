package main

//go:generate go run github.com/akavel/rsrc -manifest procguard.manifest -o rsrc.syso

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"procguard/gui"
	"procguard/internal/api"
	"procguard/internal/daemon"
	"procguard/internal/database"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

const (
	extensionId = "ilaocldmkhlifnikhinkmiepekpbefoh"
	hostName    = "com.nixuris.procguard"
)

func main() {
	// When launched by Chrome, the first argument is the extension's origin.
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "chrome-extension://") {
		gui.Run()
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
	api.StartWebServer(addr, registerWebRoutes)
}

func registerWebRoutes(srv *api.Server, r *http.ServeMux) {
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gui.HandleIndex(&srv.Mu, srv.IsAuthenticated, srv.Logger, w, r)
	})
	r.HandleFunc("/ping", gui.HandlePing)
}

func runDaemon() {
	fmt.Println("Starting ProcGuard daemon...")
	appLogger := daemon.Get()
	db, err := database.InitDB()
	if err != nil {
		appLogger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	daemon.Start(appLogger, db)
	// Keep the main goroutine alive
	select {}
}

// HandleDefaultStartup implements the main startup logic for GUI mode on Windows.
func HandleDefaultStartup() {
	// The autostart logic has been moved to a user-initiated action in the GUI.

	// Set up the native messaging host using the current executable's path.
	// Note: This means the original executable must be kept.
	// This will be improved when a proper installer is built.
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting executable path:", err)
	}
	if err := installNativeHost(exePath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install native messaging host: %v\n", err)
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
		fmt.Fprintln(os.Stderr, "Error starting API service:", err)
		return
	}

	cmdDaemon := exec.Command(exePath, "run-daemon")
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

// installNativeHost sets up the native messaging host for Chrome.
func installNativeHost(exePath string) error {
	log := daemon.Get()
	keyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + hostName

	// Check if the key already exists.
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
	if err == nil {
		k.Close()
		log.Println("Native messaging host registry key already exists.")
		return nil // Key already exists, no action needed.
	}

	log.Println("Installing native messaging host...")

	// Create the registry key.
	k, _, err = registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		log.Printf("Failed to create registry key: %v", err)
		return fmt.Errorf("failed to create registry key: %w", err)
	}
	defer k.Close()

	// Get the user cache directory.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Failed to get user cache dir: %v", err)
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}
	appDataDir := filepath.Join(cacheDir, "procguard")
	configDir := filepath.Join(appDataDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("Failed to create config directory: %v", err)
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the manifest file in the config directory.
	manifestPath := filepath.Join(configDir, "native-host.json")
	if err := createManifest(manifestPath, exePath, extensionId); err != nil {
		log.Printf("Failed to create manifest file: %v", err)
		return fmt.Errorf("failed to create manifest file: %w", err)
	}

	// Set the default value of the registry key to the manifest path.
	if err := k.SetStringValue("", manifestPath); err != nil {
		log.Printf("Failed to set registry key value: %v", err)
		return fmt.Errorf("failed to set registry key value: %w", err)
	}

	log.Println("Native messaging host installed successfully.")
	return nil
}

func createManifest(manifestPath, exePath, extensionId string) error {
	manifest := map[string]interface{}{
		"name":            hostName,
		"description":     "ProcGuard native messaging host",
		"path":            exePath,
		"type":            "stdio",
		"allowed_origins": []string{
			"chrome-extension://" + extensionId + "/",
		},
	}

	file, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}