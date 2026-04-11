package commands_test

import (
	"testing"

	"github.com/user/jira-go/internal/commands"
)

func TestTaskStatusCommandRegistered(t *testing.T) {
	cmd := commands.RootCmd
	statusCmd, _, err := cmd.Find([]string{"task", "status"})
	if err != nil || statusCmd == nil {
		t.Fatalf("task status command not registered: %v", err)
	}
}

func TestTaskStatusRequiresArgs(t *testing.T) {
	_, err := executeCommand(commands.RootCmd, "task", "status")
	if err == nil {
		t.Fatal("expected error when called with no args")
	}
}
