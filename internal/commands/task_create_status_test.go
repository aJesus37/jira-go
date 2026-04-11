package commands_test

import (
	"testing"

	"github.com/user/jira-go/internal/commands"
)

func TestTaskCreateHasStatusFlag(t *testing.T) {
	// Find the create subcommand
	createCmd, _, err := commands.RootCmd.Find([]string{"task", "create"})
	if err != nil {
		t.Fatalf("task create not found: %v", err)
	}
	if f := createCmd.Flags().Lookup("status"); f == nil {
		t.Fatal("task create --status flag not registered")
	}
}
