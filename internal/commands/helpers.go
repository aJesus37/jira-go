// internal/commands/helpers.go
package commands

import (
	"github.com/spf13/cobra"
	"github.com/aJesus37/jira-go/internal/config"
)

func getProjectKey(cmd *cobra.Command, cfg *config.Config) string {
	if projectFlag != "" {
		return projectFlag
	}
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		return project
	}
	return cfg.DefaultProject
}
