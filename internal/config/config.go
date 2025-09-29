package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "procguard", "spec.json"), nil
}

// Load reads the configuration file from the user's cache directory.
// If the file doesn't exist, it returns a default configuration.
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// If the config file doesn't exist, create a new one with default values.
		return New(), nil
	}
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		// If the file is corrupted or invalid, return an error.
		return nil, err
	}

	return &config, nil
}

// Save writes the current configuration to the configuration file.
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Marshal the config to JSON with indentation for readability.
	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write the content to the file with read/write permissions for the current user.
	return os.WriteFile(path, content, 0644)
}
