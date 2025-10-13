//go:build !windows

package app

const (
	SECURITY_MANDATORY_SYSTEM_RID = 0 // Dummy value
)

func GetProcessIntegrityLevel(pid uint32) (uint32, error) {
	// No-op on non-Windows systems. Return a value that will not cause the process to be skipped.
	// The check in process.go is `il >= SECURITY_MANDATORY_SYSTEM_RID`.
	// Returning 0 will satisfy this.
	return 0, nil
}
