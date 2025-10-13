//go:build !windows

package daemon

func EnsureAutostart() (string, error) {
	// No-op on non-Windows systems
	return "", nil
}

func RemoveAutostart() error {
	// No-op on non-Windows systems
	return nil
}
