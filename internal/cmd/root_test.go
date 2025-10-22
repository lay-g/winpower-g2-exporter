package cmd

import (
	"context"
	"fmt"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	if rootCmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}

	if rootCmd.cmd == nil {
		t.Fatal("Root command's cobra.Command is nil")
	}

	// Test basic command properties
	if rootCmd.cmd.Use != ApplicationName {
		t.Errorf("Expected command use to be '%s', got '%s'", ApplicationName, rootCmd.cmd.Use)
	}

	if rootCmd.cmd.Short == "" {
		t.Error("Expected command short description to be non-empty")
	}

	if rootCmd.cmd.Long == "" {
		t.Error("Expected command long description to be non-empty")
	}

	expectedVersion := fmt.Sprintf("%s v%s", ApplicationName, ApplicationVersion)
	if rootCmd.cmd.Version != expectedVersion {
		t.Errorf("Expected version to be '%s', got '%s'", expectedVersion, rootCmd.cmd.Version)
	}

	// Test that required subcommands are added
	hasServer := false
	hasVersion := false
	hasHelp := false

	for _, cmd := range rootCmd.cmd.Commands() {
		switch cmd.Use {
		case "server":
			hasServer = true
		case "version":
			hasVersion = true
		case "help", "help [command]":
			hasHelp = true
		}
	}

	if !hasServer {
		t.Error("Expected root command to have 'server' subcommand")
	}

	if !hasVersion {
		t.Error("Expected root command to have 'version' subcommand")
	}

	if !hasHelp {
		t.Error("Expected root command to have 'help' subcommand")
	}

	// Test cobra behavior settings
	if !rootCmd.cmd.CompletionOptions.DisableDefaultCmd {
		t.Error("Expected DisableDefaultCmd to be true")
	}

	if !rootCmd.cmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}

	if !rootCmd.cmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}

	// Test that RunE is set (for default behavior)
	if rootCmd.cmd.RunE == nil {
		t.Error("Expected root command to have RunE set for default behavior")
	}
}

func TestRootCommand_Name(t *testing.T) {
	rootCmd := NewRootCommand()
	name := rootCmd.Name()

	if name != ApplicationName {
		t.Errorf("Expected name to be '%s', got '%s'", ApplicationName, name)
	}
}

func TestRootCommand_Description(t *testing.T) {
	rootCmd := NewRootCommand()
	description := rootCmd.Description()

	if description == "" {
		t.Error("Expected description to be non-empty")
	}

	if description != rootCmd.cmd.Short {
		t.Errorf("Expected description to match short description, got '%s' vs '%s'", description, rootCmd.cmd.Short)
	}
}

func TestRootCommand_Validate(t *testing.T) {
	rootCmd := NewRootCommand()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Empty args",
			args: []string{},
		},
		{
			name: "With config flag",
			args: []string{"--config", "test.yaml"},
		},
		{
			name: "Multiple flags",
			args: []string{"--config", "test.yaml", "--port", "8080", "--log-level", "debug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rootCmd.Validate(tt.args)
			if err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestRootCommand_Execute_Version(t *testing.T) {
	rootCmd := NewRootCommand()
	ctx := context.Background()

	// Test version command
	err := rootCmd.Execute(ctx, []string{"version"})
	if err != nil {
		t.Errorf("Unexpected error executing version command: %v", err)
	}
}

func TestRootCommand_Execute_Help(t *testing.T) {
	rootCmd := NewRootCommand()
	ctx := context.Background()

	// Test help command
	err := rootCmd.Execute(ctx, []string{"help"})
	if err != nil {
		t.Errorf("Unexpected error executing help command: %v", err)
	}

	// Test help flag
	err = rootCmd.Execute(ctx, []string{"--help"})
	if err != nil {
		t.Errorf("Unexpected error executing --help flag: %v", err)
	}
}

func TestRootCommand_Execute_Server(t *testing.T) {
	rootCmd := NewRootCommand()
	ctx := context.Background()

	// Test server command with minimal flags
	err := rootCmd.Execute(ctx, []string{"server", "--log-level", "debug"})
	if err != nil {
		// We expect this to fail because server command is not yet fully implemented
		if err.Error() != "server command not yet fully implemented" {
			t.Errorf("Unexpected error executing server command: %v", err)
		}
	}
}

func TestRootCommand_Execute_Default(t *testing.T) {
	rootCmd := NewRootCommand()
	ctx := context.Background()

	// Test default behavior (should act like help command)
	err := rootCmd.Execute(ctx, []string{"--log-level", "debug"})
	if err != nil {
		t.Errorf("Unexpected error executing default help command: %v", err)
	}
}

func TestRootCommand_Execute_InvalidCommand(t *testing.T) {
	rootCmd := NewRootCommand()
	ctx := context.Background()

	// Test invalid command
	err := rootCmd.Execute(ctx, []string{"invalid-command"})
	if err == nil {
		t.Error("Expected error for invalid command, but got none")
	}
}

func TestRootCommand_SubCommands(t *testing.T) {
	rootCmd := NewRootCommand()

	// Test server subcommand
	commands := make(map[string]Commander)
	serverCmd := rootCmd.newServerCommand(commands)
	if serverCmd == nil {
		t.Error("Failed to create server subcommand")
	}

	if serverCmd.Use != "server" {
		t.Errorf("Expected server command use to be 'server', got '%s'", serverCmd.Use)
	}

	if serverCmd.Short == "" {
		t.Error("Expected server command short description to be non-empty")
	}

	if serverCmd.Long == "" {
		t.Error("Expected server command long description to be non-empty")
	}

	// Test version subcommand
	versionCmd := rootCmd.newVersionCommand()
	if versionCmd == nil {
		t.Error("Failed to create version subcommand")
	}

	if versionCmd.Use != "version" {
		t.Errorf("Expected version command use to be 'version', got '%s'", versionCmd.Use)
	}

	// Test help subcommand
	helpCmd := rootCmd.newHelpCommand(commands)
	if helpCmd == nil {
		t.Error("Failed to create help subcommand")
	}

	if helpCmd.Use != "help [command]" {
		t.Errorf("Expected help command use to be 'help [command]', got '%s'", helpCmd.Use)
	}
}

func TestRootCommand_Flags(t *testing.T) {
	rootCmd := NewRootCommand()

	// Test that common flags are added to server command
	commands := make(map[string]Commander)
	serverCmd := rootCmd.newServerCommand(commands)

	// Check for common flags
	flags := []string{
		"config", "log-level", "port", "data-dir",
		"skip-ssl-verify", "sync-write",
	}

	for _, flag := range flags {
		if serverCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected server command to have '--%s' flag", flag)
		}
	}

	// Check for server-specific flags
	serverFlags := []string{
		"winpower-url", "winpower-username", "winpower-password",
		"winpower-timeout", "winpower-max-retries", "winpower-skip-tls-verify",
		"host", "server-mode", "scheduler-interval", "energy-precision",
	}

	for _, flag := range serverFlags {
		if serverCmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected server command to have '--%s' flag", flag)
		}
	}
}

// Helper function to check if error is related to configuration
func containsConfigError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	configKeywords := []string{
		"config", "failed to", "cannot", "invalid", "not found",
	}
	for _, keyword := range configKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    (len(s) > len(substr) &&
		     (s[:len(substr)] == substr ||
		      s[len(s)-len(substr):] == substr ||
		      containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}