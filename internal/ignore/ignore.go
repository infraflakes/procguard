package ignore

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

	// Other user-level services
	"fcitx5-",
	"music-discord-rpc",

	// Browser/Electron helpers
	"Web Content",
	"Isolated Web Co",
	"Socket Process",
	"Privileged Cont",
	"RDD Process",
	"WebExtensions",
	"Utility Process",
	"crashhelper",
	"forkserver",
	"MainThread",
	".zen-twilight-w",
	".ghostty-wrappe",
	".wl-copy-wrappe",

	// Other
	"Xwayland",
	"ssh-agent",
}

// IsIgnored checks if a process name should be ignored based on the ignore list.
// It performs both exact and prefix matching, and handles truncated names.
func IsIgnored(name string, ignoreList []string) bool {
	// Handle truncated names that start with a dot, like ".gvfsd-http-wr"
	nameToCompare := name
	if strings.HasPrefix(nameToCompare, ".") {
		nameToCompare = strings.TrimPrefix(nameToCompare, ".")
	}

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