package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"procguard/internal/data"

	"golang.org/x/sys/windows/registry"
)

const (
	HostName = "com.nixuris.procguard"
)

// InstallNativeHost sets up the native messaging host for Chrome.
func InstallNativeHost(exePath, extensionId string) error {
	log := data.GetLogger()
	keyPath := `SOFTWARE\Google\Chrome\NativeMessagingHosts\` + HostName

	log.Printf("Attempting to install native messaging host for extension ID: %s", extensionId)

	// Create the registry key.
	k, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		log.Printf("Failed to create registry key: %v", err)
		return fmt.Errorf("failed to create registry key: %w", err)
	}
	defer func() {
		if err := k.Close(); err != nil {
			log.Printf("Failed to close registry key: %v", err)
		}
	}()

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
	if err := CreateManifest(manifestPath, exePath, extensionId); err != nil {
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

func CreateManifest(manifestPath, exePath, extensionId string) error {
	manifest := map[string]interface{}{
		"name":        HostName,
		"description": "ProcGuard native messaging host",
		"path":        exePath,
		"type":        "stdio",
		"allowed_origins": []string{
			"chrome-extension://" + extensionId + "/",
		},
	}

	file, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			data.GetLogger().Printf("Failed to close file: %v", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(manifest)
}

func RegisterExtension(extensionId string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}
	return InstallNativeHost(exePath, extensionId)
}
