// internal/commands/init.go
package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/aJesus37/jira-go/internal/config"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize jira configuration",
	Long:  `Interactive setup to configure jira with your Jira instance.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config if available
	existingCfg, _ := config.Load()
	defaultProject := ""
	if existingCfg != nil {
		defaultProject = existingCfg.DefaultProject
	}

	fmt.Println("🚀 Welcome to jira!")
	if existingCfg != nil {
		fmt.Println("Editing existing configuration. Press Enter to keep current values.")
	}
	fmt.Println("Let's set up your configuration.")
	fmt.Println()

	// Jira URL
	jiraURL := getInputWithDefault(reader, "Jira URL", getExistingJiraURL(existingCfg, defaultProject))

	// Email
	email := getInputWithDefault(reader, "Email", getExistingEmail(existingCfg))

	// API Token (hidden input)
	apiToken := getPasswordInput(reader, existingCfg)

	// Default Project
	projectKey := getInputWithDefault(reader, "Default Project Key", defaultProject)
	if projectKey == "" {
		projectKey = "SCRUM" // fallback
	}

	// Board ID
	boardID := getExistingBoardID(existingCfg, defaultProject)
	boardIDStr := getInputWithDefault(reader, "Board ID", formatIntDefault(boardID))
	if boardIDStr != "" {
		fmt.Sscanf(boardIDStr, "%d", &boardID)
	}

	// Multi-owner field (optional)
	multiOwnerField := getInputWithDefault(reader, "Multi-owner Custom Field ID (optional)", getExistingMultiOwnerField(existingCfg, defaultProject))

	// Sprint field (optional)
	sprintField := getInputWithDefault(reader, "Sprint Custom Field ID (optional)", getExistingSprintField(existingCfg, defaultProject))

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
				BoardID:         boardID,
				MultiOwnerField: multiOwnerField,
				SprintField:     sprintField,
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

// Helper functions to get existing values

func getExistingJiraURL(cfg *config.Config, projectKey string) string {
	if cfg == nil {
		return ""
	}
	if project, ok := cfg.Projects[projectKey]; ok {
		return project.JiraURL
	}
	return ""
}

func getExistingEmail(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	return cfg.Auth.Email
}

func getExistingBoardID(cfg *config.Config, projectKey string) int {
	if cfg == nil {
		return 0
	}
	if project, ok := cfg.Projects[projectKey]; ok {
		return project.BoardID
	}
	return 0
}

func getExistingMultiOwnerField(cfg *config.Config, projectKey string) string {
	if cfg == nil {
		return ""
	}
	if project, ok := cfg.Projects[projectKey]; ok {
		return project.MultiOwnerField
	}
	return ""
}

func getExistingSprintField(cfg *config.Config, projectKey string) string {
	if cfg == nil {
		return ""
	}
	if project, ok := cfg.Projects[projectKey]; ok {
		return project.SprintField
	}
	return ""
}

func formatIntDefault(val int) string {
	if val == 0 {
		return ""
	}
	return fmt.Sprintf("%d", val)
}

func getInputWithDefault(reader *bufio.Reader, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func getPasswordInput(reader *bufio.Reader, cfg *config.Config) string {
	if cfg != nil && cfg.Auth.APIToken != "" {
		fmt.Print("API Token (input hidden) [keep current]: ")
	} else {
		fmt.Print("API Token (input hidden): ")
	}

	apiTokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println()
		if cfg != nil {
			return cfg.Auth.APIToken
		}
		return ""
	}
	apiToken := string(apiTokenBytes)
	fmt.Println()

	if apiToken == "" && cfg != nil {
		return cfg.Auth.APIToken
	}
	return apiToken
}

func init() {
	rootCmd.AddCommand(initCmd)
}
