package version

import "fmt"

// Build metadata injected at link time via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// String returns the user-facing version line.
func String() string {
	if Commit == "unknown" && Date == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}

// Short returns the semver tag only.
func Short() string {
	return Version
}
