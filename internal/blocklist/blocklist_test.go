package blocklist

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadSave(t *testing.T) {

	tempDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tempDir)

	// Test saving a blocklist
	listToSave := []string{"a", "b", "c"}
	if err := Save(listToSave); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Test loading the saved blocklist
	loadedList, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !reflect.DeepEqual(loadedList, listToSave) {
		t.Errorf("Load() got = %v, want %v", loadedList, listToSave)
	}

	// Test that an empty list is loaded when the file doesn't exist
	if err := os.Remove(filepath.Join(tempDir, "procguard", blockListFile)); err != nil {
		t.Fatalf("failed to remove blocklist file: %v", err)
	}

	emptyList, err := Load()
	if err != nil {
		t.Fatalf("Load() error after removing file = %v", err)
	}
	if len(emptyList) != 0 {
		t.Errorf("Load() after removing file should be empty, got %v", emptyList)
	}
}
