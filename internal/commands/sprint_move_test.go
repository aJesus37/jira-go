package commands_test

import (
	"testing"

	"github.com/aJesus37/jira-go/internal/commands"
)

func TestSprintMoveHasFromSprintFlag(t *testing.T) {
	cmd, _, err := commands.RootCmd.Find([]string{"sprint", "move"})
	if err != nil || cmd == nil {
		t.Fatalf("sprint move not found: %v", err)
	}
	if f := cmd.Flags().Lookup("from-sprint"); f == nil {
		t.Fatal("sprint move --from-sprint flag not registered")
	}
}

func TestSprintMoveRequiresAtLeastTargetID(t *testing.T) {
	_, err := executeCommand(commands.RootCmd, "sprint", "move")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}
