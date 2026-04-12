package commands

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var (
	updateCheckOnly bool
	updateYes       bool
	updatePathFlag  string
)

type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update jira to the latest version",
	Long: `Check for a newer version on GitHub Releases and update the binary.

By default, installs to the user-local directory (~/.local/bin on Linux/macOS,
%LOCALAPPDATA%\Programs\jira on Windows). Use --path to override the destination.

Exit codes when using --check:
  0 = already up to date
  1 = update available`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if version == "dev" {
		return fmt.Errorf("cannot update a dev build — install a tagged release first")
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("could not reach GitHub API: %w", err)
	}

	currentVersion := normalizeVersion(version)
	latestVersion := normalizeVersion(release.TagName)

	if currentVersion == latestVersion {
		fmt.Printf("Already up to date (%s)\n", latestVersion)
		return nil
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Printf("Latest version:  %s\n", latestVersion)

	if updateCheckOnly {
		fmt.Println("Update available.")
		os.Exit(1)
	}

	if !updateYes {
		fmt.Print("Update? [y/N] ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	assetName := determineAssetName(runtime.GOOS, runtime.GOARCH)
	downloadURL := ""
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (looking for %s)",
			runtime.GOOS, runtime.GOARCH, assetName)
	}

	installDir := updatePathFlag
	if installDir == "" {
		installDir, err = getInstallDirectory()
		if err != nil {
			return fmt.Errorf("determining install directory: %w", err)
		}
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("creating install directory: %w", err)
	}

	if err := downloadAndInstall(downloadURL, installDir, runtime.GOOS); err != nil {
		return fmt.Errorf("installing update: %w", err)
	}

	fmt.Printf("Updated to %s\n", latestVersion)
	if !isInPATH(installDir) {
		fmt.Printf("Note: %s is not in your PATH\n", installDir)
	}
	return nil
}

func fetchLatestRelease() (*githubRelease, error) {
	req, err := http.NewRequest("GET",
		"https://api.github.com/repos/aJesus37/jira-go/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &release, nil
}

func downloadAndInstall(url, installDir, goos string) error {
	tmpFile, err := os.CreateTemp("", "jira-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() { _ = os.Remove(tmpName) }()

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		_ = tmpFile.Close()
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing download: %w", err)
	}
	_ = tmpFile.Close()

	binaryName := "jira"
	if goos == "windows" {
		binaryName = "jira.exe"
	}
	destPath := filepath.Join(installDir, binaryName)

	if goos == "windows" {
		return extractZip(tmpName, binaryName, destPath)
	}
	return extractTarGz(tmpName, binaryName, destPath)
}

func extractTarGz(archivePath, binaryName, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close() //nolint:errcheck

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}
		if filepath.Base(hdr.Name) == binaryName {
			return writeExecutable(tr, destPath)
		}
	}
	return fmt.Errorf("binary %q not found in archive", binaryName)
}

func extractZip(archivePath, binaryName, destPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close() //nolint:errcheck

	for _, f := range r.File {
		if filepath.Base(f.Name) == binaryName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close() //nolint:errcheck
			return writeExecutable(rc, destPath)
		}
	}
	return fmt.Errorf("binary %q not found in archive", binaryName)
}

func writeExecutable(r io.Reader, destPath string) error {
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".jira-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName) // no-op if rename succeeded
	}()

	if err := tmp.Chmod(0755); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}
	if _, err := io.Copy(tmp, r); err != nil {
		return fmt.Errorf("writing binary: %w", err)
	}
	tmp.Close()

	return os.Rename(tmpName, destPath)
}

func determineAssetName(goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("jira_%s_%s%s", goos, goarch, ext)
}

func normalizeVersion(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false,
		"Check for updates without downloading (exit 1 if update available)")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false,
		"Skip confirmation prompt")
	updateCmd.Flags().StringVar(&updatePathFlag, "path", "",
		"Override install directory")
}
