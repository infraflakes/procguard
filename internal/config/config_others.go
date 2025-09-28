//go:build !windows

package config

// Config defines the structure of the application's configuration file.
// It is used to store state and settings that need to persist between runs.
type Config struct {
	// SystemdInstalled tracks whether the systemd service has been installed.
	SystemdInstalled bool `json:"systemd_installed"`
	// PasswordHash stores the bcrypt hash of the GUI password.
	PasswordHash string `json:"password_hash,omitempty"`
}

// New creates a new Config with default values.
func New() *Config {
	return &Config{SystemdInstalled: false, PasswordHash: ""}
}
