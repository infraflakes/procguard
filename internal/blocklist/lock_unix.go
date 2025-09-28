//go:build linux || darwin

package blocklist

// platformLock is a no-op on Unix (0600 already set).
func platformLock(path string) error { return nil }
