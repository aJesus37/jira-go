package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/aJesus37/jira-go/skills"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage Claude Code skills bundled with jira-go",
}

var skillsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install bundled Claude Code skills",
	Long: `Install the Claude Code skills bundled in this binary.

Skills are written to ~/.agents/skills/jira-go/ and symlinked into
~/.claude/skills/ so Claude Code can discover them automatically.`,
	RunE: runSkillsInstall,
}

func runSkillsInstall(_ *cobra.Command, _ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	agentsDir := filepath.Join(homeDir, ".agents", "skills", "jira-go")
	claudeSkillsDir := filepath.Join(homeDir, ".claude", "skills")

	if err := os.RemoveAll(agentsDir); err != nil {
		return fmt.Errorf("clearing existing skills: %w", err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("creating skills directory: %w", err)
	}
	if err := os.MkdirAll(claudeSkillsDir, 0755); err != nil {
		return fmt.Errorf("creating Claude skills directory: %w", err)
	}

	err = fs.WalkDir(skills.FS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		dest := filepath.Join(agentsDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, readErr := skills.FS.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(dest, data, 0644)
	})
	if err != nil {
		return fmt.Errorf("extracting skills: %w", err)
	}

	fmt.Printf("Installed skills to %s\n", agentsDir)

	entries, err := fs.ReadDir(skills.FS, ".")
	if err != nil {
		return fmt.Errorf("reading skills: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(agentsDir, entry.Name())
		linkPath := filepath.Join(claudeSkillsDir, entry.Name())
		_ = os.Remove(linkPath)
		if err := os.Symlink(skillPath, linkPath); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", entry.Name(), err)
		}
		fmt.Printf("Linked  %s\n", linkPath)
	}

	fmt.Println("\nSkills installed. Claude Code will pick them up automatically.")
	return nil
}

func init() {
	rootCmd.AddCommand(skillsCmd)
	skillsCmd.AddCommand(skillsInstallCmd)
}
