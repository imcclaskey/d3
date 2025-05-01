package version

// Version is the current i3 version
const Version = "0.1.0"

// VersionWithPrefix returns the version with 'v' prefix
func VersionWithPrefix() string {
	return "v" + Version
} 