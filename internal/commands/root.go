// internal/commands/root.go
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/aJesus37/jira-go/internal/config"
)

var (
	// Global flags
	projectFlag       string
	verboseFlag       bool
	noInteractiveFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "jira",
	Short: "A CLI tool for managing Jira Software projects",
	Long: `jira is a comprehensive CLI for Jira Software that supports
task management, sprint operations, epics, and agile ceremonies.

By default, commands run in interactive TUI mode.
Use --no-interactive flag for automation/AI agent compatibility.`,
	RunE: runDefault,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip for init command
		if cmd.Name() == "init" || cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}
		return nil
	},
}

// runDefault determines what to do when jira-go is run without subcommands
func runDefault(cmd *cobra.Command, args []string) error {
	// Check if config exists
	configPath := config.GetConfigPath()
	_, err := os.Stat(configPath)

	if os.IsNotExist(err) {
		// No config found, run init
		fmt.Println("No configuration found. Let's set up jira!")
		fmt.Println()
		return runInit(cmd, args)
	}

	// Config exists, run task list
	return runTaskList(cmd, args)
}

// RootCmd exposes the root command for testing.
var RootCmd = rootCmd

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Project key (overrides config)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noInteractiveFlag, "no-interactive", false, "Disable interactive TUI mode (for automation/CI)")
}
