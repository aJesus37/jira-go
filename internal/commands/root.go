// internal/commands/root.go
package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jira-go",
	Short: "A CLI tool for managing Jira Software projects",
	Long: `jira-go is a comprehensive CLI for Jira Software that supports
task management, sprint operations, epics, and agile ceremonies.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// TODO: Add subcommands
}
