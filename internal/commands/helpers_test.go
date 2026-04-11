package commands_test

import (
	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (string, error) {
	root.SetArgs(args)
	_, err := root.ExecuteC()
	return "", err
}
