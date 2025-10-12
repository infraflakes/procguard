package daemon

import "strings"

// DefaultLinux is the default list of user-level process names to ignore on Linux.
var DefaultLinux = []string{
	// Systemd processes (user-level)
	"systemd",
	"(sd-pam)",

	// DBus and other daemons
	"dbus-daemon",

	// GNOME services
	".gcr-ssh-agent-",
	".gnome-keyring-",
	".gnome-session-",
	".gnome-shell-wr",
	".at-spi-bus-lau",
	"at-spi2-registryd",
	".gnome-shell-ca",
	"dconf-service",
	".evolution-sour",
	".org.gnome.Shel",
	".evolution-alar",
	".org.gnome.Scre",
	".goa-daemon-wra",
	".goa-identity-s",
	".evolution-cale",
	".evolution-addr",
	"gsd-",   // gnome-settings-daemon prefix
	"gvfsd-", // gnome-virtual-file-system prefix
	"gvfs-",
	"gdm-",

	// XDG and desktop portals
	"xdg-",
	"fusermount3",
	".mutter-x11-fra",
	".localsearch-3-",

	// Pipewire and audio
	"pipewire",
	"pipewire-pulse",
	"wireplumber",
	"speech-dispatcher",

	// Other
	"Xwayland",
	"ssh-agent",
}

// DefaultWindows is the default list of process names to ignore on Windows.
var DefaultWindows = []string{
	// Core system processes that might run at Medium IL and are safe to ignore.
	"System Idle Process",
	"System",
	"smss.exe",
	"lsass.exe",
	"StartMenuExperienceHost.exe",
	"SearchHost.exe",
	"RuntimeBroker.exe",
	"svchost.exe",
	"sihost.exe",
	"ShellHost.exe",
	"taskhostw.exe",
	"conhost.exe",
	"dllhost.exe",
	"LockApp.exe",
	"CHXSmartScreen.exe",
	"smartscreen.exe",
	"ipf_helper.exe",
	"TextInputHost.exe",
	"RtkUWP.exe",
	"RtkAudUService64.exe",
	"MoNotificationUx.exe",
	"ctfmon.exe",
	"nvcontainer.exe",
	"ShellExperienceHost.exe",
	"audiodg.exe",
	"ApplicationFrameHost.exe",
	"backgroundTaskHost.exe",
	"SystemSettings.exe",
	"wininit.exe",
	"fontdrvhost.exe",
	"dwm.exe",
	"spoolsv.exe",
	"services.exe", // Parent process check should handle most children, but we can ignore the parent itself.
}

// IsIgnored checks if a process name should be ignored based on the ignore list.
// It performs both exact and prefix matching, and handles truncated names.
func IsIgnored(name string, ignoreList []string) bool {
	// Handle truncated names that start with a dot, like ".gvfsd-http-wr"
	nameToCompare := strings.TrimPrefix(name, ".")

	for _, ignored := range ignoreList {
		if strings.HasSuffix(ignored, "-") {
			// Prefix match (e.g., "gsd-" should match "gsd-color")
			if strings.HasPrefix(name, strings.TrimSuffix(ignored, "-")) {
				return true
			}
			// Also check against the dot-stripped name for truncated processes
			if strings.HasPrefix(nameToCompare, strings.TrimSuffix(ignored, "-")) {
				return true
			}
		} else {
			// Exact match
			if name == ignored {
				return true
			}
		}
	}
	return false
}
