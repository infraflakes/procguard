//go:build !windows

package data

func platformLock(path string) error {
	// No-op on non-Windows systems
	return nil
}
