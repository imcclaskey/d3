package version

// Version is the current i3 version
const Version = "0.0.0"

// VersionWithPrefix returns the version with 'v' prefix
func VersionWithPrefix() string {
	return "v" + Version
} 