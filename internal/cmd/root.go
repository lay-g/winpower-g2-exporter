package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// ApplicationName represents the name of the application
	ApplicationName = "winpower-g2-exporter"
	// ApplicationVersion represents the current version of the application (injected at build time)
	ApplicationVersion = "dev"
)

// Commander interface defines the contract for command execution
type Commander interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args []string) error
	Validate(args []string) error
}

// RootCommand represents the root command of the application
type RootCommand struct {
	cmd *cobra.Command
}

// NewRootCommand creates a new root command with all subcommands
func NewRootCommand() *RootCommand {
	rootCmd := &RootCommand{
		cmd: &cobra.Command{
			Use:   ApplicationName,
			Short: "WinPower G2 Exporter - Prometheus metrics exporter",
			Long: `WinPower G2 Exporter is a Prometheus metrics exporter that collects data
from WinPower devices, calculates energy consumption, and exports metrics.`,
			Version: fmt.Sprintf("%s v%s", ApplicationName, ApplicationVersion),
		},
	}

	// Initialize commands map for help command
	commands := make(map[string]Commander)

	// Create Commander instances
	serverCommand := NewServerCommand(DefaultCLIArgs())
	versionCommand := NewVersionCommand()

	// Create cobra commands
	serverCmd := rootCmd.newServerCommand(commands)
	versionCmd := rootCmd.newVersionCommand()
	helpCmd := rootCmd.newHelpCommand(commands)

	// Add Commander instances to map
	commands["server"] = serverCommand
	commands["version"] = versionCommand

	// Add subcommands to cobra
	rootCmd.cmd.AddCommand(serverCmd)
	rootCmd.cmd.AddCommand(versionCmd)
	rootCmd.cmd.AddCommand(helpCmd)

	// Update help command with commands map
	helpCmdInstance := NewHelpCommand(commands)
	helpCmdInstance.SetCommands(commands)

	// Set help as default command
	rootCmd.cmd.SetVersionTemplate("{{.Name}} {{printf \"%s\" .Version}}\n")

	// Configure cobra behavior
	rootCmd.cmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.cmd.SilenceUsage = true
	rootCmd.cmd.SilenceErrors = true

	// Add common flags to root command for default behavior
	rootCmd.addCommonFlags(rootCmd.cmd)

	// If no command is specified, show help using Cobra's built-in help system
	rootCmd.cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Use Cobra's built-in help system instead of our custom help command
		return cmd.Help()
	}

	return rootCmd
}

// Execute executes the root command with the given arguments
func (r *RootCommand) Execute(ctx context.Context, args []string) error {
	r.cmd.SetArgs(args)
	return r.cmd.ExecuteContext(ctx)
}

// Name returns the name of the root command
func (r *RootCommand) Name() string {
	return r.cmd.Use
}

// Description returns the description of the root command
func (r *RootCommand) Description() string {
	return r.cmd.Short
}

// Validate validates the root command arguments
func (r *RootCommand) Validate(args []string) error {
	// Basic validation can be added here if needed
	return nil
}

// GetCLIArgs extracts and returns parsed CLI arguments
func (r *RootCommand) GetCLIArgs() *CLIArgs {
	// This will be implemented by individual commands
	// For now, return default args
	return DefaultCLIArgs()
}

// newServerCommand creates the server subcommand
func (r *RootCommand) newServerCommand(commands map[string]Commander) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the HTTP server (default command)",
		Long: `Start the HTTP server that exposes Prometheus metrics.
This is the default command and will be executed if no subcommand is specified.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse command line flags
			flagParser := NewFlagParser(cmd)
			cliArgs, err := flagParser.ParseFlags()
			if err != nil {
				return fmt.Errorf("failed to parse command line flags: %w", err)
			}

			// Execute server command
			serverCommand := NewServerCommand(cliArgs)
			return serverCommand.Execute(cmd.Context(), args)
		},
	}

	// Add server-specific flags
	r.addCommonFlags(cmd)
	r.addServerFlags(cmd)

	return cmd
}

// newVersionCommand creates the version subcommand
func (r *RootCommand) newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display detailed version information including build time and Git commit.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Execute version command
			versionCommand := NewVersionCommand()
			return versionCommand.Execute(cmd.Context(), args)
		},
	}

	return cmd
}

// newHelpCommand creates the help subcommand
func (r *RootCommand) newHelpCommand(commands map[string]Commander) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Show help information",
		Long:  "Display help information for any command or show general usage information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Execute our custom help command for specific functionality
			helpCommand := NewHelpCommand(commands)
			return helpCommand.Execute(cmd.Context(), args)
		},
	}

	return cmd
}

// addCommonFlags adds common flags that are shared across commands
func (r *RootCommand) addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "./config.yaml", "Configuration file path")
	cmd.Flags().StringP("log-level", "l", "info", "Log level (debug|info|warn|error)")
	cmd.Flags().IntP("port", "p", 9090, "HTTP service port")
	cmd.Flags().String("data-dir", "./data", "Data directory path")
	cmd.Flags().Bool("skip-ssl-verify", false, "Skip SSL certificate verification")
	cmd.Flags().Bool("sync-write", true, "Enable synchronous writes")
}

// addServerFlags adds server-specific flags
func (r *RootCommand) addServerFlags(cmd *cobra.Command) {
	// WinPower module flags
	cmd.Flags().String("winpower-url", "", "WinPower server URL")
	cmd.Flags().String("winpower-username", "", "WinPower username")
	cmd.Flags().String("winpower-password", "", "WinPower password")
	cmd.Flags().Duration("winpower-timeout", 0, "WinPower request timeout")
	cmd.Flags().Int("winpower-max-retries", 0, "WinPower maximum retry attempts")
	cmd.Flags().Bool("winpower-skip-tls-verify", false, "Skip TLS verification for WinPower")

	// Server module flags
	cmd.Flags().String("host", "", "Server host address")
	cmd.Flags().String("server-mode", "", "Server mode (debug|release|test)")

	// Scheduler module flags
	cmd.Flags().Duration("scheduler-interval", 0, "Scheduler collection interval")

	// Energy module flags
	cmd.Flags().Float64("energy-precision", 0, "Energy calculation precision")
}