package version

// These variables are populated at build time using -ldflags
var (
	// Version is the semantic version of the application
	Version = "dev"

	// BuildTime is the time the binary was built
	BuildTime = "unknown"
)

// GetVersion returns the current version of the application
func GetVersion() string {
	return Version
}

// GetBuildTime returns the build time of the binary
func GetBuildTime() string {
	return BuildTime
}

// GetVersionInfo returns a formatted string with version information
func GetVersionInfo() string {
	return "Hisame v" + Version + " (built " + BuildTime + ")"
}
