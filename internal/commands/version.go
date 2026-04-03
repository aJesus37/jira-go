// internal/commands/version.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// These are set during build
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("jira version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  build date: %s\n", buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
