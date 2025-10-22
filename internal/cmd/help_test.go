package cmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock command for testing
type mockCommand struct {
	name        string
	description string
}

func (m *mockCommand) Name() string {
	return m.name
}

func (m *mockCommand) Description() string {
	return m.description
}

func (m *mockCommand) Execute(ctx context.Context, args []string) error {
	return nil
}

func (m *mockCommand) Validate(args []string) error {
	return nil
}

func TestNewHelpCommand(t *testing.T) {
	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

	assert.NotNil(t, cmd)
	assert.Equal(t, commands, cmd.GetCommands())
	assert.Equal(t, "help", cmd.Name())
	assert.Equal(t, "Show help information", cmd.Description())
}

func TestHelpCommand_Validate(t *testing.T) {
	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

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
			name:      "valid single arg",
			args:      []string{"server"},
			expectErr: false,
		},
		{
			name:      "invalid multiple args",
			args:      []string{"server", "extra"},
			expectErr: true,
			errMsg:    "help command takes at most one argument",
		},
		{
			name:      "invalid too many args",
			args:      []string{"arg1", "arg2", "arg3"},
			expectErr: true,
			errMsg:    "help command takes at most one argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestHelpCommand_Execute(t *testing.T) {
	commands := make(map[string]Commander)
	serverCmd := &mockCommand{name: "server", description: "Start the server"}
	versionCmd := &mockCommand{name: "version", description: "Show version"}

	commands["server"] = serverCmd
	commands["version"] = versionCmd

	cmd := NewHelpCommand(commands)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "general help",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "command help",
			args:      []string{"server"},
			expectErr: false,
		},
		{
			name:      "too many args",
			args:      []string{"server", "extra"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestHelpCommand_SetCommands(t *testing.T) {
	cmd := NewHelpCommand(make(map[string]Commander))

	// Initially empty
	assert.Empty(t, cmd.GetCommands())

	// Set new commands
	newCommands := make(map[string]Commander)
	newCommands["test"] = &mockCommand{name: "test", description: "Test command"}

	cmd.SetCommands(newCommands)

	assert.Equal(t, newCommands, cmd.GetCommands())
	assert.Len(t, cmd.GetCommands(), 1)
	assert.Contains(t, cmd.GetCommands(), "test")
}

func TestHelpCommand_GetCommands(t *testing.T) {
	commands := make(map[string]Commander)
	commands["server"] = &mockCommand{name: "server", description: "Start the server"}
	commands["version"] = &mockCommand{name: "version", description: "Show version"}

	cmd := NewHelpCommand(commands)

	retrievedCommands := cmd.GetCommands()
	assert.Equal(t, commands, retrievedCommands)
	assert.Len(t, retrievedCommands, 2)
	assert.Contains(t, retrievedCommands, "server")
	assert.Contains(t, retrievedCommands, "version")
}

func TestHelpCommand_PrintCommandHelp(t *testing.T) {
	commands := make(map[string]Commander)
	serverCmd := &mockCommand{name: "server", description: "Start the HTTP server"}

	commands["server"] = serverCmd

	cmd := NewHelpCommand(commands)

	// This test is mainly for code coverage as printCommandHelp writes to stdout
	// We'll just make sure it doesn't panic
	assert.NotPanics(t, func() {
		cmd.printCommandHelp("server")
	})

	// Test with unknown command
	assert.NotPanics(t, func() {
		cmd.printCommandHelp("unknown")
	})
}

func TestHelpCommand_PrintGeneralHelp(t *testing.T) {
	commands := make(map[string]Commander)
	commands["server"] = &mockCommand{name: "server", description: "Start the HTTP server"}
	commands["version"] = &mockCommand{name: "version", description: "Show version information"}

	cmd := NewHelpCommand(commands)

	// This test is mainly for code coverage as printGeneralHelp writes to stdout
	// We'll just make sure it doesn't panic
	assert.NotPanics(t, func() {
		cmd.printGeneralHelp()
	})
}

func TestHelpCommand_PrintServerCommandHelp(t *testing.T) {
	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

	// This test is mainly for code coverage as printServerCommandHelp writes to stdout
	// We'll just make sure it doesn't panic
	assert.NotPanics(t, func() {
		cmd.printServerCommandHelp()
	})
}

func TestHelpCommand_PrintVersionCommandHelp(t *testing.T) {
	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

	// This test is mainly for code coverage as printVersionCommandHelp writes to stdout
	// We'll just make sure it doesn't panic
	assert.NotPanics(t, func() {
		cmd.printVersionCommandHelp()
	})
}

func TestHelpCommand_PrintHelpCommandHelp(t *testing.T) {
	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

	// This test is mainly for code coverage as printHelpCommandHelp writes to stdout
	// We'll just make sure it doesn't panic
	assert.NotPanics(t, func() {
		cmd.printHelpCommandHelp()
	})
}

func TestHelpCommand_UnknownCommandHandling(t *testing.T) {
	commands := make(map[string]Commander)
	commands["server"] = &mockCommand{name: "server", description: "Start the server"}

	cmd := NewHelpCommand(commands)

	// Test that unknown command doesn't panic and shows available commands
	assert.NotPanics(t, func() {
		cmd.printCommandHelp("nonexistent")
	})
}

func TestHelpCommand_CommanderInterface(t *testing.T) {
	var _ Commander = (*HelpCommand)(nil)

	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

	assert.Equal(t, "help", cmd.Name())
	assert.Equal(t, "Show help information", cmd.Description())
}

func TestHelpCommand_WithEmptyCommands(t *testing.T) {
	// Test help command with empty commands map
	commands := make(map[string]Commander)
	cmd := NewHelpCommand(commands)

	// Should handle empty commands gracefully
	assert.NotPanics(t, func() {
		cmd.printGeneralHelp()
	})

	assert.NotPanics(t, func() {
		cmd.printCommandHelp("server")
	})
}

func TestHelpCommand_WithNilCommands(t *testing.T) {
	// Create help command with nil commands
	cmd := &HelpCommand{
		commands: nil,
	}

	// Should handle nil commands gracefully (though NewHelpCommand prevents this)
	assert.NotPanics(t, func() {
		cmd.Validate([]string{})
	})

	assert.NotPanics(t, func() {
		cmd.printCommandHelp("server")
	})
}