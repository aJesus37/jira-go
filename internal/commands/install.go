// internal/commands/install.go
package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sys/execabs"
)

var (
	installPathFlag   string
	installForceFlag  bool
	installGlobalFlag bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install jira binary to system PATH",
	Long: `Install the jira binary to a directory in your system PATH.

By default, installs to a user-local directory that doesn't require admin privileges:
  - Windows: %LOCALAPPDATA%\Programs\jira\
  - macOS/Linux: ~/.local/bin

Use --global to install system-wide (may require admin privileges):
  - Windows: C:\Program Files\jira\
  - macOS/Linux: /usr/local/bin

The installer will automatically elevate permissions if required.`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Get the current executable path
	currentExec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate current executable: %w", err)
	}

	// Resolve to absolute path
	currentExec, err = filepath.EvalSymlinks(currentExec)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	// Determine install directory
	installDir, err := getInstallDirectory()
	if err != nil {
		return fmt.Errorf("cannot determine install directory: %w", err)
	}

	// If custom path specified, use it
	if installPathFlag != "" {
		installDir = installPathFlag
	}

	// Create install directory if needed
	if err := os.MkdirAll(installDir, 0755); err != nil {
		// Try with elevated permissions
		if err := mkdirElevated(installDir); err != nil {
			return fmt.Errorf("cannot create install directory %s: %w", installDir, err)
		}
	}

	// Determine target path
	targetPath := filepath.Join(installDir, getBinaryName())

	// Check if already exists
	if _, err := os.Stat(targetPath); err == nil && !installForceFlag {
		return fmt.Errorf("binary already exists at %s (use --force to overwrite)", targetPath)
	}

	// Copy binary
	if err := copyBinary(currentExec, targetPath); err != nil {
		// Try with elevated permissions
		if err := copyBinaryElevated(currentExec, targetPath); err != nil {
			return fmt.Errorf("cannot install binary: %w", err)
		}
	}

	// Verify installation
	if !isInPATH(installDir) {
		fmt.Printf("\n⚠️  Warning: %s is not in your PATH\n", installDir)
		fmt.Println("\nAdd it to your PATH:")

		switch runtime.GOOS {
		case "windows":
			fmt.Printf("  [System Environment Variables] → Add to PATH: %s\n", installDir)
		default:
			shell := getShell()
			switch shell {
			case "zsh":
				fmt.Printf("  echo 'export PATH=\"%s:$PATH\"' >> ~/.zshrc\n", installDir)
			case "bash":
				fmt.Printf("  echo 'export PATH=\"%s:$PATH\"' >> ~/.bashrc\n", installDir)
			default:
				fmt.Printf("  Add to your shell profile: export PATH=\"%s:$PATH\"\n", installDir)
			}
		}
	}

	fmt.Printf("\n✅ Successfully installed jira to: %s\n", targetPath)
	fmt.Printf("\nYou can now use 'jira' from anywhere!\n")

	return nil
}

// getInstallDirectory returns the appropriate install directory based on OS and flags
func getInstallDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get user home directory: %w", err)
	}

	if installGlobalFlag {
		// System-wide installation
		switch runtime.GOOS {
		case "windows":
			return `C:\Program Files\jira`, nil
		case "darwin":
			return "/usr/local/bin", nil
		default: // linux and others
			return "/usr/local/bin", nil
		}
	}

	// User-local installation (no admin required)
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		return filepath.Join(localAppData, "Programs", "jira"), nil
	case "darwin":
		// macOS: prefer ~/.local/bin, fallback to ~/bin
		localBin := filepath.Join(homeDir, ".local", "bin")
		if _, err := os.Stat(localBin); err == nil {
			return localBin, nil
		}
		return filepath.Join(homeDir, "bin"), nil
	default: // linux and others
		// Prefer XDG_DATA_HOME, fallback to ~/.local/bin
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome != "" {
			return filepath.Join(dataHome, "..", "bin"), nil
		}
		return filepath.Join(homeDir, ".local", "bin"), nil
	}
}

// getBinaryName returns the binary name with appropriate extension
func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "jira.exe"
	}
	return "jira"
}

// isInPATH checks if a directory is in the system PATH
func isInPATH(dir string) bool {
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	for _, p := range paths {
		// Normalize paths for comparison
		if filepath.Clean(p) == filepath.Clean(dir) {
			return true
		}
	}
	return false
}

// getShell returns the current shell name
func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}
	return filepath.Base(shell)
}

// copyBinary copies the binary from source to destination
func copyBinary(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Remove existing file if present
	os.Remove(dst)

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// mkdirElevated creates a directory with elevated permissions
func mkdirElevated(path string) error {
	switch runtime.GOOS {
	case "windows":
		// Use PowerShell to create directory with elevation
		psCmd := fmt.Sprintf(`Start-Process powershell -Verb runAs -ArgumentList '-Command New-Item -ItemType Directory -Force -Path "%s"' -Wait`, path)
		cmd := execabs.Command("powershell", "-Command", psCmd)
		return cmd.Run()
	case "darwin", "linux":
		// Try sudo
		cmd := execabs.Command("sudo", "mkdir", "-p", path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("elevated permissions not supported on this platform")
	}
}

// copyBinaryElevated copies the binary with elevated permissions
func copyBinaryElevated(src, dst string) error {
	switch runtime.GOOS {
	case "windows":
		// Use PowerShell to copy with elevation
		psCmd := fmt.Sprintf(`Start-Process powershell -Verb runAs -ArgumentList '-Command Copy-Item -Force -Path "%s" -Destination "%s"' -Wait`, src, dst)
		cmd := execabs.Command("powershell", "-Command", psCmd)
		return cmd.Run()
	case "darwin", "linux":
		// Try sudo
		cmd := execabs.Command("sudo", "cp", src, dst)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Also set executable permissions
		if err := cmd.Run(); err != nil {
			return err
		}

		chmodCmd := execabs.Command("sudo", "chmod", "755", dst)
		chmodCmd.Stdin = os.Stdin
		chmodCmd.Stdout = os.Stdout
		chmodCmd.Stderr = os.Stderr
		return chmodCmd.Run()
	default:
		return fmt.Errorf("elevated permissions not supported on this platform")
	}
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringVarP(&installPathFlag, "path", "p", "", "Custom installation path")
	installCmd.Flags().BoolVarP(&installForceFlag, "force", "f", false, "Overwrite existing binary")
	installCmd.Flags().BoolVarP(&installGlobalFlag, "global", "g", false, "Install system-wide (may require admin privileges)")
}
