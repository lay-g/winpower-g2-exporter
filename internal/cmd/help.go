package cmd

import (
	"context"
	"fmt"
	"strings"
)

// HelpCommand implements the help subcommand
type HelpCommand struct {
	commands map[string]Commander
}

// NewHelpCommand creates a new help command instance
func NewHelpCommand(commands map[string]Commander) *HelpCommand {
	return &HelpCommand{
		commands: commands,
	}
}

// Name returns the command name
func (h *HelpCommand) Name() string {
	return "help"
}

// Description returns the command description
func (h *HelpCommand) Description() string {
	return "Show help information"
}

// Validate validates the help command arguments
func (h *HelpCommand) Validate(args []string) error {
	// Help command can accept 0 or 1 argument
	if len(args) > 1 {
		return fmt.Errorf("help command takes at most one argument")
	}
	return nil
}

// Execute executes the help command
func (h *HelpCommand) Execute(ctx context.Context, args []string) error {
	if err := h.Validate(args); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(args) == 0 {
		h.printGeneralHelp()
	} else {
		h.printCommandHelp(args[0])
	}

	return nil
}

// printGeneralHelp prints general help information
func (h *HelpCommand) printGeneralHelp() {
	fmt.Printf(`%s - WinPower G2 Exporter

DESCRIPTION:
  WinPower G2 Exporter is a Prometheus metrics exporter that collects data
  from WinPower devices, calculates energy consumption, and exports metrics.

USAGE:
  %s [command] [flags]

COMMANDS:
`, ApplicationName, ApplicationName)

	// Print available commands
	maxNameLength := 0
	for name := range h.commands {
		if len(name) > maxNameLength {
			maxNameLength = len(name)
		}
	}

	for name, cmd := range h.commands {
		padding := strings.Repeat(" ", maxNameLength-len(name)+2)
		fmt.Printf("  %s%s%s\n", name, padding, cmd.Description())
	}

	fmt.Printf(`
FLAGS:
%s

EXAMPLES:
  # Start the server with default configuration
  %s server

  # Start the server with custom configuration file
  %s server --config /path/to/config.yaml

  # Start the server with specific port and log level
  %s server --port 8080 --log-level debug

  # Show version information
  %s version

  # Show help for a specific command
  %s help server

CONFIGURATION EXAMPLE:
  # Basic configuration file (config.yaml)
  storage:
    data_dir: "./data"
    sync_write: true

  winpower:
    url: "https://winpower.example.com:8080"
    username: "admin"
    password: "password"
    timeout: "30s"
    skip_ssl_verify: false

  server:
    port: 9090
    host: "0.0.0.0"
    mode: "release"

  scheduler:
    collection_interval: "5s"

  energy:
    precision: 0.01

ENVIRONMENT VARIABLES:
  WINPOWER_EXPORTER_CONFIG        Configuration file path
  WINPOWER_EXPORTER_LOG_LEVEL     Log level (debug|info|warn|error)
  WINPOWER_EXPORTER_PORT          HTTP service port
  WINPOWER_EXPORTER_DATA_DIR      Data directory path
  WINPOWER_EXPORTER_SKIP_SSL_VERIFY Skip SSL certificate verification
  WINPOWER_EXPORTER_SYNC_WRITE    Enable synchronous writes

  WinPower module:
  WINPOWER_EXPORTER_WINPOWER_URL      WinPower server URL
  WINPOWER_EXPORTER_WINPOWER_USERNAME WinPower username
  WINPOWER_EXPORTER_WINPOWER_PASSWORD WinPower password

  Server module:
  WINPOWER_EXPORTER_HOST        Server host address
  WINPOWER_EXPORTER_SERVER_MODE Server mode (debug|release|test)

  Scheduler module:
  WINPOWER_EXPORTER_SCHEDULER_INTERVAL Scheduler collection interval

  Energy module:
  WINPOWER_EXPORTER_ENERGY_PRECISION Energy calculation precision

For more information about a specific command, use:
  %s help [command]

`, GetFlagHelp(), ApplicationName, ApplicationName, ApplicationName, ApplicationName, ApplicationName, ApplicationName)
}

// printCommandHelp prints help for a specific command
func (h *HelpCommand) printCommandHelp(commandName string) {
	cmd, exists := h.commands[commandName]
	if !exists {
		fmt.Printf("Unknown command: %s\n\n", commandName)
		fmt.Printf("Available commands:\n")
		for name := range h.commands {
			fmt.Printf("  %s\n", name)
		}
		fmt.Printf("\nUse '%s help' to see general help information.\n", ApplicationName)
		return
	}

	fmt.Printf(`%s - %s

USAGE:
  %s %s

DESCRIPTION:
  %s

`, strings.ToUpper(commandName), cmd.Description(), ApplicationName, commandName, cmd.Description())

	// Add command-specific help based on the command type
	switch commandName {
	case "server":
		h.printServerCommandHelp()
	case "version":
		h.printVersionCommandHelp()
	case "help":
		h.printHelpCommandHelp()
	}

	fmt.Printf(`FLAGS:
%s
`, GetFlagHelp())
}

// printServerCommandHelp prints specific help for the server command
func (h *HelpCommand) printServerCommandHelp() {
	fmt.Printf(`SERVER COMMAND DETAILS:

The server command starts the HTTP server and begins collecting data from WinPower devices.
This is the default command and will be executed if no subcommand is specified.

MODULE INITIALIZATION ORDER:
  1. Storage     - Initializes data storage and file management
  2. Auth        - Sets up authentication and token management
  3. Energy      - Configures energy calculation engine
  4. Collector   - Initializes data collection from WinPower devices
  5. Metrics     - Sets up Prometheus metrics registry
  6. Server      - Starts HTTP server with metrics endpoint
  7. Scheduler   - Starts periodic data collection

SIGNAL HANDLING:
  The server responds to SIGTERM and SIGINT signals for graceful shutdown.
  During shutdown, modules are stopped in reverse initialization order.

HEALTH ENDPOINTS:
  /health      - Basic health check
  /metrics     - Prometheus metrics endpoint
  /debug/pprof - Debug profiling (if enabled)

`)
}

// printVersionCommandHelp prints specific help for the version command
func (h *HelpCommand) printVersionCommandHelp() {
	fmt.Printf(`VERSION COMMAND DETAILS:

The version command displays detailed version information including:
- Application version number
- Build timestamp
- Git commit hash
- Go runtime version
- Target platform (OS/architecture)

VERSION FORMAT:
  The version follows semantic versioning (MAJOR.MINOR.PATCH).
  Development builds are marked as "dev".

BUILD INFORMATION:
  Build information is injected during compilation using build flags.
  The VERSION file in the project root contains the version number.

`)
}

// printHelpCommandHelp prints specific help for the help command
func (h *HelpCommand) printHelpCommandHelp() {
	fmt.Printf(`HELP COMMAND DETAILS:

The help command displays usage information for the application or specific commands.

USAGE EXAMPLES:
  %s help           # Show general help
  %s help server    # Show help for server command
  %s help version   # Show help for version command

`, ApplicationName, ApplicationName, ApplicationName)
}

// SetCommands updates the available commands
func (h *HelpCommand) SetCommands(commands map[string]Commander) {
	h.commands = commands
}

// GetCommands returns the available commands
func (h *HelpCommand) GetCommands() map[string]Commander {
	return h.commands
}