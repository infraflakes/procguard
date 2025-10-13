//go:build !windows

package web

func InstallNativeHost(exePath string) error {
	// No-op on non-Windows systems
	return nil
}
