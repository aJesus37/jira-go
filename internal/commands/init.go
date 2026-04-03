// internal/commands/init.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/user/jira-go/internal/config"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize jira-go configuration",
	Long:  `Interactive setup to configure jira-go with your Jira instance.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Welcome to jira-go!")
	fmt.Println("Let's set up your configuration.")
	fmt.Println()

	// Jira URL
	fmt.Print("Jira URL (e.g., https://your-domain.atlassian.net): ")
	jiraURL, _ := reader.ReadString('\n')
	jiraURL = strings.TrimSpace(jiraURL)

	// Email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// API Token (hidden input)
	fmt.Print("API Token (input hidden): ")
	apiTokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("reading API token: %w", err)
	}
	apiToken := string(apiTokenBytes)
	fmt.Println()

	// Default Project
	fmt.Print("Default Project Key (e.g., PROJ): ")
	projectKey, _ := reader.ReadString('\n')
	projectKey = strings.TrimSpace(projectKey)

	// Multi-owner field (optional)
	fmt.Print("Multi-owner Custom Field ID (optional, e.g., customfield_10001): ")
	multiOwnerField, _ := reader.ReadString('\n')
	multiOwnerField = strings.TrimSpace(multiOwnerField)

	// Create config
	cfg := &config.Config{
		DefaultProject: projectKey,
		Auth: config.AuthConfig{
			Email:    email,
			APIToken: apiToken,
		},
		Projects: map[string]config.Project{
			projectKey: {
				JiraURL:         jiraURL,
				MultiOwnerField: multiOwnerField,
			},
		},
		Cache: config.CacheConfig{
			Enabled:    true,
			DefaultTTL: "30m",
		},
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("\n✓ Configuration saved to %s\n", config.GetConfigPath())
	fmt.Println("✓ You're ready to use jira-go!")
	fmt.Println("\nTry: jira-go --help")

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
