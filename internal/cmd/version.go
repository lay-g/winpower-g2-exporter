package cmd

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Version information (will be injected at build time)
var (
	// Version represents the application version
	Version = "dev"
	// BuildTime represents the build timestamp
	BuildTime = "unknown"
	// GitCommit represents the git commit hash
	GitCommit = "unknown"
	// GoVersion represents the Go version used for build
	GoVersion = runtime.Version()
)

// VersionInfo holds comprehensive version information
type VersionInfo struct {
	Version   string    `json:"version"`
	BuildTime time.Time `json:"build_time"`
	GitCommit string    `json:"git_commit"`
	GoVersion string    `json:"go_version"`
	Platform  string    `json:"platform"`
}

// VersionCommand implements the version subcommand
type VersionCommand struct{}

// NewVersionCommand creates a new version command instance
func NewVersionCommand() *VersionCommand {
	return &VersionCommand{}
}

// Name returns the command name
func (v *VersionCommand) Name() string {
	return "version"
}

// Description returns the command description
func (v *VersionCommand) Description() string {
	return "Show version information"
}

// Validate validates the version command arguments
func (v *VersionCommand) Validate(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("version command takes no arguments")
	}
	return nil
}

// Execute executes the version command
func (v *VersionCommand) Execute(ctx context.Context, args []string) error {
	if err := v.Validate(args); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	versionInfo := v.getVersionInfo()
	v.printVersionInfo(versionInfo)

	return nil
}

// getVersionInfo returns comprehensive version information
func (v *VersionCommand) getVersionInfo() *VersionInfo {
	buildTime, err := time.Parse("2006-01-02_15:04:05", BuildTime)
	if err != nil {
		// If parsing fails, use zero time
		buildTime = time.Time{}
	}

	return &VersionInfo{
		Version:   Version,
		BuildTime: buildTime,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// printVersionInfo prints version information in a user-friendly format
func (v *VersionCommand) printVersionInfo(info *VersionInfo) {
	// Remove 'v' prefix if present to avoid duplication
	version := strings.TrimPrefix(info.Version, "v")
	fmt.Printf("%s version %s\n", ApplicationName, version)

	// Display additional version details
	fmt.Printf("Build Time: %s\n", v.formatBuildTime(info.BuildTime))
	fmt.Printf("Git Commit: %s\n", v.formatGitCommit(info.GitCommit))
	fmt.Printf("Go Version: %s\n", info.GoVersion)
	fmt.Printf("Platform: %s\n", info.Platform)
}

// formatBuildTime formats the build time for display
func (v *VersionCommand) formatBuildTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04:05 UTC")
}

// formatGitCommit formats the git commit hash for display
func (v *VersionCommand) formatGitCommit(commit string) string {
	if commit == "unknown" || commit == "" {
		return "unknown"
	}
	// If the commit hash is longer than 7 characters, truncate it
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

// GetVersionInfoJSON returns version information in JSON format
func (v *VersionCommand) GetVersionInfoJSON() *VersionInfo {
	return v.getVersionInfo()
}

// IsDevelopmentVersion checks if the current version is a development version
func (v *VersionCommand) IsDevelopmentVersion() bool {
	return Version == "dev" || Version == ""
}

// GetShortVersion returns a short version string suitable for logging
func (v *VersionCommand) GetShortVersion() string {
	if v.IsDevelopmentVersion() {
		return fmt.Sprintf("%s-dev", ApplicationName)
	}
	return fmt.Sprintf("%s-v%s", ApplicationName, Version)
}

// CompareVersions compares two version strings (simple implementation)
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (v *VersionCommand) CompareVersions(v1, v2 string) int {
	// This is a simplified version comparison
	// In a production environment, you might want to use a more sophisticated
	// version comparison library that handles semantic versioning properly

	if v1 == v2 {
		return 0
	}

	if v1 < v2 {
		return -1
	}

	return 1
}