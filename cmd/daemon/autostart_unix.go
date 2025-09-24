//go:build linux || darwin

package daemon

func checkAutostart() {
	// This is a no-op on non-Windows systems.
}
