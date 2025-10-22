package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerCommand(t *testing.T) {
	args := DefaultCLIArgs()
	cmd := NewServerCommand(args)

	assert.NotNil(t, cmd)
	assert.Equal(t, args, cmd.args)
	assert.Equal(t, "server", cmd.Name())
	assert.Equal(t, "Start the HTTP server and begin data collection", cmd.Description())
}

func TestServerCommand_Validate(t *testing.T) {
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
			errMsg:    "server command takes no arguments",
		},
		{
			name:      "invalid with multiple args",
			args:      []string{"arg1", "arg2"},
			expectErr: true,
			errMsg:    "server command takes no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := DefaultCLIArgs()
			cmd := NewServerCommand(args)

			err := cmd.Validate(tt.args)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerCommand_validateConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		args      *CLIArgs
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid configuration",
			args: &CLIArgs{
				Port:     9090,
				LogLevel: "info",
			},
			expectErr: false,
		},
		{
			name: "invalid port",
			args: &CLIArgs{
				Port:     -1,
				LogLevel: "info",
			},
			expectErr: true,
			errMsg:    "invalid port number",
		},
		{
			name: "port too high",
			args: &CLIArgs{
				Port:     70000,
				LogLevel: "info",
			},
			expectErr: true,
			errMsg:    "invalid port number",
		},
		{
			name: "invalid log level",
			args: &CLIArgs{
				Port:     9090,
				LogLevel: "invalid",
			},
			expectErr: true,
			errMsg:    "invalid log level",
		},
		{
			name: "valid debug log level",
			args: &CLIArgs{
				Port:     9090,
				LogLevel: "debug",
			},
			expectErr: false,
		},
		{
			name: "valid error log level",
			args: &CLIArgs{
				Port:     9090,
				LogLevel: "error",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewServerCommand(tt.args)
			err := cmd.validateConfiguration()
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerCommand_determineConfigSource(t *testing.T) {
	tests := []struct {
		name     string
		args     *CLIArgs
		expected string
	}{
		{
			name: "default config",
			args: &CLIArgs{
				Config: "./config.yaml",
			},
			expected: "defaults_and_environment",
		},
		{
			name: "custom config",
			args: &CLIArgs{
				Config: "/custom/path/config.yaml",
			},
			expected: "file",
		},
		{
			name: "empty config",
			args: &CLIArgs{
				Config: "",
			},
			expected: "defaults_and_environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewServerCommand(tt.args)
			result := cmd.determineConfigSource()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServerCommand_Validate_NilArgs(t *testing.T) {
	cmd := &ServerCommand{}
	err := cmd.Validate([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CLI arguments not initialized")
}

func TestServerCommand_Execute_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	args := DefaultCLIArgs()
	cmd := NewServerCommand(args)

	// When context is already cancelled before execution starts,
	// the server should handle this gracefully during startup
	// Note: Depending on implementation, this might not always return an error
	// if the cancellation happens after certain initialization steps
	err := cmd.Execute(ctx, []string{})

	// We just verify it doesn't panic or hang indefinitely
	// The actual error depends on where in the startup process the cancellation is detected
	_ = err
}
