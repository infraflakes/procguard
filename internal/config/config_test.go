package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tempDir)

	// Test saving a config
	configToSave := New()
	configToSave.PasswordHash = "myhash"
	if err := configToSave.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test loading the saved config
	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !reflect.DeepEqual(loadedConfig, configToSave) {
		t.Errorf("Load() got = %v, want %v", loadedConfig, configToSave)
	}

	// Test that a new config is loaded when the file doesn't exist
	if err := os.Remove(filepath.Join(tempDir, "procguard", "spec.json")); err != nil {
		t.Fatalf("failed to remove config file: %v", err)
	}

	newConfig, err := Load()
	if err != nil {
		t.Fatalf("Load() error after removing file = %v", err)
	}

	if !reflect.DeepEqual(newConfig, New()) {
		t.Errorf("Load() after removing file got = %v, want %v", newConfig, New())
	}
}
