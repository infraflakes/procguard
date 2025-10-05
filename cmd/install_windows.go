package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"procguard/internal/logger"

	"golang.org/x/sys/windows/registry"
)

const extensionId = "ilaocldmkhlifnikhinkmiepekpbefoh"
const hostName = "com.nixuris.procguard"

// installNativeHost sets up the native messaging host for Chrome.
func installNativeHost() error {
	log := logger.Get()
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

	// Get the path to the executable.
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get the user cache directory.
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Failed to get user cache dir: %v", err)
		return fmt.Errorf("failed to get user cache dir: %w", err)
	}
	appDataDir := filepath.Join(cacheDir, "procguard")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		log.Printf("Failed to create app data directory: %v", err)
		return fmt.Errorf("failed to create app data directory: %w", err)
	}

	// Create the manifest file in the app data directory.
	manifestPath := filepath.Join(appDataDir, "procguard.json")
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
		"name":              hostName,
		"description":       "ProcGuard native messaging host",
		"path":              exePath,
		"type":              "stdio",
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
