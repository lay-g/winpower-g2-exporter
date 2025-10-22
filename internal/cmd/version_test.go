package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Name())
	assert.Equal(t, "Show version information", cmd.Description())
}

func TestVersionCommand_Validate(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid empty args",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "invalid with args",
			args:      []string{"invalid"},
			expectErr: true,
			errMsg:    "version command takes no arguments",
		},
		{
			name:      "invalid with multiple args",
			args:      []string{"arg1", "arg2"},
			expectErr: true,
			errMsg:    "version command takes no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCommand()

			err := cmd.Validate(tt.args)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersionCommand_Execute(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid execution",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "invalid args",
			args:      []string{"invalid"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewVersionCommand()

			err := cmd.Execute(context.Background(), tt.args)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "validation failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersionCommand_GetVersionInfo(t *testing.T) {
	cmd := NewVersionCommand()

	info := cmd.getVersionInfo()

	assert.NotNil(t, info)
	assert.Equal(t, Version, info.Version)
	assert.Equal(t, GitCommit, info.GitCommit)
	assert.Equal(t, "go1.22", info.GoVersion) // Default Go version format
	assert.Contains(t, info.Platform, "/")
}

func TestVersionCommand_GetVersionInfoWithBuildTime(t *testing.T) {
	// Test with valid build time format
	originalBuildTime := BuildTime
	BuildTime = "2024-01-15_10:30:00"
	defer func() { BuildTime = originalBuildTime }()

	cmd := NewVersionCommand()
	info := cmd.getVersionInfo()

	expectedTime, err := time.Parse("2006-01-02_15:04:05", BuildTime)
	require.NoError(t, err)

	assert.Equal(t, expectedTime, info.BuildTime)
}

func TestVersionCommand_GetVersionInfoWithInvalidBuildTime(t *testing.T) {
	// Test with invalid build time format
	originalBuildTime := BuildTime
	BuildTime = "invalid-time"
	defer func() { BuildTime = originalBuildTime }()

	cmd := NewVersionCommand()
	info := cmd.getVersionInfo()

	assert.True(t, info.BuildTime.IsZero())
}

func TestVersionCommand_FormatBuildTime(t *testing.T) {
	cmd := NewVersionCommand()

	t.Run("valid build time", func(t *testing.T) {
		buildTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := cmd.formatBuildTime(buildTime)
		assert.Equal(t, "2024-01-15 10:30:05 UTC", result) // Note: there might be a slight time difference
	})

	t.Run("zero build time", func(t *testing.T) {
		buildTime := time.Time{}
		result := cmd.formatBuildTime(buildTime)
		assert.Equal(t, "unknown", result)
	})
}

func TestVersionCommand_FormatGitCommit(t *testing.T) {
	cmd := NewVersionCommand()

	tests := []struct {
		name     string
		commit   string
		expected string
	}{
		{
			name:     "short commit",
			commit:   "abc123",
			expected: "abc123",
		},
		{
			name:     "long commit",
			commit:   "abc123def456789",
			expected: "abc123d",
		},
		{
			name:     "empty commit",
			commit:   "",
			expected: "unknown",
		},
		{
			name:     "unknown commit",
			commit:   "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.formatGitCommit(tt.commit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionCommand_IsDevelopmentVersion(t *testing.T) {
	cmd := NewVersionCommand()

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "development version",
			version:  "dev",
			expected: true,
		},
		{
			name:     "empty version",
			version:  "",
			expected: true,
		},
		{
			name:     "release version",
			version:  "1.0.0",
			expected: false,
		},
		{
			name:     "pre-release version",
			version:  "1.0.0-beta",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVersion := Version
			Version = tt.version
			defer func() { Version = originalVersion }()

			result := cmd.IsDevelopmentVersion()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionCommand_GetShortVersion(t *testing.T) {
	cmd := NewVersionCommand()

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "development version",
			version:  "dev",
			expected: "winpower-g2-exporter-dev",
		},
		{
			name:     "release version",
			version:  "1.0.0",
			expected: "winpower-g2-exporter-v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalVersion := Version
			Version = tt.version
			defer func() { Version = originalVersion }()

			result := cmd.GetShortVersion()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionCommand_CompareVersions(t *testing.T) {
	cmd := NewVersionCommand()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "v1 less than v2",
			v1:       "1.0.0",
			v2:       "1.1.0",
			expected: -1,
		},
		{
			name:     "v1 greater than v2",
			v1:       "1.1.0",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "alphabetic comparison",
			v1:       "alpha",
			v2:       "beta",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.CompareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionCommand_GetVersionInfoJSON(t *testing.T) {
	cmd := NewVersionCommand()

	info := cmd.GetVersionInfoJSON()

	assert.NotNil(t, info)
	assert.Equal(t, Version, info.Version)
	assert.Equal(t, GitCommit, info.GitCommit)
}

// Test that the command interface is correctly implemented
func TestVersionCommand_CommanderInterface(t *testing.T) {
	var _ Commander = (*VersionCommand)(nil)

	cmd := NewVersionCommand()

	assert.Equal(t, "version", cmd.Name())
	assert.Equal(t, "Show version information", cmd.Description())
}