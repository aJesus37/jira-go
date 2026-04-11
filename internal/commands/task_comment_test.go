package commands_test

import (
	"testing"
	"github.com/user/jira-go/internal/commands"
)

func TestTaskCommentCommandRegistered(t *testing.T) {
	commentCmd, _, err := commands.RootCmd.Find([]string{"task", "comment"})
	if err != nil || commentCmd == nil {
		t.Fatalf("task comment command not registered: %v", err)
	}
}

func TestTaskCommentRequiresTwoArgs(t *testing.T) {
	commands.RootCmd.SetArgs([]string{"task", "comment", "PROJ-1"})
	err := commands.RootCmd.Execute()
	if err == nil {
		t.Fatal("expected error with only one arg (missing message)")
	}
}
