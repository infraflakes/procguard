package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	SystemdInstalled bool `json:"systemd_installed"`
}

func GetConfigPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "procguard", "spec.json"), nil
}

func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return a default config if the file doesn't exist
		return &Config{SystemdInstalled: false}, nil
	}
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}
