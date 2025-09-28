//go:build windows

package config

// Config defines the structure of the application's configuration file.
// It is used to store state and settings that need to persist between runs.
type Config struct {
	// AutostartEnabled tracks whether the Windows autostart task has been created.
	AutostartEnabled bool `json:"autostart_enabled"`
	// PasswordHash stores the bcrypt hash of the GUI password.
	PasswordHash string `json:"password_hash,omitempty"`
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{AutostartEnabled: false, PasswordHash: ""}
}
