// internal/commands/root.go
package commands

import (
	"github.com/spf13/cobra"
)

var (
	// Global flags
	projectFlag  string
	noCacheFlag  bool
	cacheTTLFlag string
	verboseFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "jira-go",
	Short: "A CLI tool for managing Jira Software projects",
	Long: `jira-go is a comprehensive CLI for Jira Software that supports
task management, sprint operations, epics, and agile ceremonies.

Use "jira-go init" to get started with the initial configuration.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip for init command
		if cmd.Name() == "init" || cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}

		// TODO: Initialize cache based on flags
		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&projectFlag, "project", "p", "", "Project key (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&noCacheFlag, "no-cache", false, "Disable cache for this command")
	rootCmd.PersistentFlags().StringVar(&cacheTTLFlag, "cache-ttl", "", "Cache TTL (e.g., 5m, 1h)")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
}
