# OSS Release Readiness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make jira-go a proper open source project with a correct module path, CI quality gate, GoReleaser cross-platform release pipeline, and a `jira update` self-update command.

**Architecture:** Four independent workstreams executed in sequence: (1) fix stale references and docs, (2) add GitHub Actions CI, (3) add GoReleaser + release workflow, (4) add `jira update` command that downloads from GitHub Releases into the user-local bin directory.

**Tech Stack:** Go 1.21, GoReleaser v2, GitHub Actions, Cobra CLI, stdlib `archive/tar` + `archive/zip` + `compress/gzip`

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `go.mod` | Modify | Fix module path + Go version |
| `internal/**/*.go` | Modify | Bulk-replace import paths |
| `internal/commands/cache.go` | Modify | Fix import path |
| `README.md` | Modify | Fix URLs, remove stale flags, add `u` shortcut |
| `CONTRIBUTING.md` | Modify | Fix repo URL |
| `CHANGELOG.md` | Modify | Version unreleased section, fix URLs |
| `.github/workflows/ci.yml` | Create | Test + lint gate on push/PR |
| `.github/workflows/release.yml` | Create | GoReleaser trigger on `v*` tags |
| `.goreleaser.yaml` | Create | Cross-platform build config |
| `internal/commands/update.go` | Create | `jira update` command |
| `internal/commands/update_test.go` | Create | Unit tests for pure functions |

---

## Task 1: Fix Module Path, Go Version, and Docs

**Files:**
- Modify: `go.mod`
- Modify: all `internal/**/*.go` (bulk sed)
- Modify: `README.md`
- Modify: `CONTRIBUTING.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Update go.mod module path and Go version**

Open `go.mod`. Change the first two lines from:
```
module github.com/user/jira-go

go 1.25.7
```
to:
```
module github.com/aJesus37/jira-go

go 1.21
```

- [ ] **Step 2: Bulk-replace import paths in all Go files**

```bash
find . -name "*.go" -not -path "./.git/*" | xargs sed -i 's|github.com/user/jira-go|github.com/aJesus37/jira-go|g'
```

- [ ] **Step 3: Replace placeholder URLs in docs**

```bash
sed -i 's|github.com/user/jira-go|github.com/aJesus37/jira-go|g' README.md CONTRIBUTING.md CHANGELOG.md
sed -i 's|github.com/user/jira|github.com/aJesus37/jira-go|g' README.md CONTRIBUTING.md
```

- [ ] **Step 4: Verify the build compiles**

```bash
go mod tidy
go build ./...
```

Expected: no errors. If `go mod tidy` upgrades the `go` directive above 1.21 (because a dependency requires it), that's fine — accept the change.

- [ ] **Step 5: Run tests to verify nothing broke**

```bash
go test ./...
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum
git add $(git diff --name-only)
git commit -m "chore: fix module path to github.com/aJesus37/jira-go and set go 1.21"
```

---

## Task 2: README Corrections

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Fix the Installation > From Source section**

Find the current block:
```markdown
### From Source

```bash
git clone https://github.com/user/jira
cd jira
task build
```
```

Replace with:
```markdown
### From Source

```bash
git clone https://github.com/aJesus37/jira-go
cd jira-go
task build
```
```

- [ ] **Step 2: Fix the Download Binary releases page reference**

Find:
```markdown
Download the latest release from the releases page, then run:
```

Replace with:
```markdown
Download the latest release from the [releases page](https://github.com/aJesus37/jira-go/releases), then run:
```

- [ ] **Step 3: Remove stale `--no-cache` and `--cache-ttl` flags from Global Flags section**

Find and remove these two lines from the Global Flags list:
```markdown
- `--no-cache` - Disable cache for this command
- `--cache-ttl` - Custom cache TTL (e.g., `5m`, `1h`)
```

After removal the Global Flags section should read:
```markdown
## Global Flags

All commands support these global flags:

- `-p, --project` - Override default project
- `-v, --verbose` - Enable verbose output
```

- [ ] **Step 4: Add owner management shortcut to Sprint Board keyboard shortcuts table**

Find the keyboard shortcuts table. Add a row for `u` after the `c` row:

```markdown
| `u` | Manage owners (add/remove) |
```

Full table after edit:
```markdown
| Key | Action |
|-----|--------|
| `←/→` or `h/l` | Navigate between columns |
| `↑/↓` or `k/j` | Navigate tickets in column |
| `x` | Toggle column visibility (hide/show) |
| `+/-` | Resize column width |
| `f` | Toggle focus between visible and hidden columns |
| `[` / `]` | Move column left/right (reorder) |
| `d` or `Enter` | View ticket details |
| `s` | Change ticket status |
| `u` | Manage owners (add/remove) |
| `c` | Add comment |
| `q` or `Esc` | Quit |
```

- [ ] **Step 5: Add `jira update` to the README**

Add a new section after the Cache Management section:

```markdown
### Update

```bash
# Check for updates
jira update --check

# Update to latest version
jira update

# Update without confirmation
jira update --yes
```
```

- [ ] **Step 6: Commit**

```bash
git add README.md
git commit -m "docs: fix README URLs, remove stale flags, add update and owner shortcuts"
```

---

## Task 3: Version the CHANGELOG

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add entries from the last two merged PRs to Unreleased**

The CHANGELOG's `[Unreleased]` section is missing entries from recent work. Add these to the `### Added` list under `[Unreleased]`:

```markdown
- Interactive owner management in TUI (`u` key): add/remove owners with live user search
- `jira update` command: self-update binary from GitHub Releases
- `--format` flag defaults to `table` across task, sprint, and report commands
- Sprint date validation: end date cannot be before start date
- Jira URL scheme validation on config load
- Dynamic standup status ordering and coloring in ceremony view
- `SearchIssuesAll` API method for automatic pagination
- JQL injection protection via `escapeJQL` helper
```

- [ ] **Step 2: Rename `[Unreleased]` to `[0.2.0]` with today's date**

Change:
```markdown
## [Unreleased]
```
to:
```markdown
## [Unreleased]

## [0.2.0] - 2026-04-11
```

- [ ] **Step 3: Update comparison URLs at the bottom**

Find:
```markdown
[Unreleased]: https://github.com/user/jira-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/user/jira-go/releases/tag/v0.1.0
```

Replace with:
```markdown
[Unreleased]: https://github.com/aJesus37/jira-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/aJesus37/jira-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/aJesus37/jira-go/releases/tag/v0.1.0
```

- [ ] **Step 4: Commit**

```bash
git add CHANGELOG.md
git commit -m "chore: version unreleased section as v0.2.0 in CHANGELOG"
```

---

## Task 4: CI Workflow

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Create the GitHub Actions CI workflow**

```bash
mkdir -p .github/workflows
```

Create `.github/workflows/ci.yml` with this exact content:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Verify go.mod is tidy
        run: |
          go mod tidy
          git diff --exit-code go.mod go.sum

      - name: Verify formatting
        run: |
          go fmt ./...
          git diff --exit-code

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./...
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions CI workflow (test, lint, fmt, mod-tidy checks)"
```

---

## Task 5: GoReleaser Config

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create GoReleaser configuration**

Create `.goreleaser.yaml` at the repo root:

```yaml
version: 2

project_name: jira

before:
  hooks:
    - go mod tidy

builds:
  - id: jira
    main: ./cmd/jira
    binary: jira
    ldflags:
      - -s -w
      - -X github.com/aJesus37/jira-go/internal/commands.version={{.Version}}
      - -X github.com/aJesus37/jira-go/internal/commands.commit={{.Commit}}
      - -X github.com/aJesus37/jira-go/internal/commands.buildDate={{.Date}}
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: linux
        goarch: arm64
      - goos: windows
        goarch: arm64

archives:
  - id: default
    name_template: "jira_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        formats: [zip]
    files:
      - LICENSE
      - README.md

checksum:
  name_template: checksums.txt
  algorithm: sha256

changelog:
  sort: asc
  use: github

release:
  github:
    owner: aJesus37
    name: jira-go
```

- [ ] **Step 2: Verify GoReleaser config is valid (dry run)**

Install GoReleaser if not present:
```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

Run a snapshot build to verify the config:
```bash
goreleaser build --snapshot --clean
```

Expected: builds succeed for all 4 targets, archives appear in `dist/`.

- [ ] **Step 3: Clean up dist and commit**

```bash
rm -rf dist/
git add .goreleaser.yaml
git commit -m "chore: add GoReleaser config for cross-platform releases"
```

---

## Task 6: Release Workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create the release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add GoReleaser release workflow triggered on v* tags"
```

---

## Task 7: `jira update` — Pure Function Tests (TDD)

**Files:**
- Create: `internal/commands/update_test.go`
- Create (skeleton): `internal/commands/update.go`

- [ ] **Step 1: Write failing tests for `determineAssetName` and `normalizeVersion`**

Create `internal/commands/update_test.go`:

```go
package commands

import (
	"testing"
)

func TestDetermineAssetName(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"linux", "amd64", "jira_linux_amd64.tar.gz"},
		{"darwin", "amd64", "jira_darwin_amd64.tar.gz"},
		{"darwin", "arm64", "jira_darwin_arm64.tar.gz"},
		{"windows", "amd64", "jira_windows_amd64.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			got := determineAssetName(tt.goos, tt.goarch)
			if got != tt.want {
				t.Errorf("determineAssetName(%q, %q) = %q, want %q",
					tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0.2.0", "v0.2.0"},
		{"v0.2.0", "v0.2.0"},
		{"1.0.0", "v1.0.0"},
		{"v1.0.0", "v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeVersion(tt.input)
			if got != tt.want {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
go test ./internal/commands/... -run "TestDetermineAssetName|TestNormalizeVersion" -v
```

Expected: `FAIL` — `determineAssetName` and `normalizeVersion` are not defined yet.

- [ ] **Step 3: Create `update.go` with just the two pure functions**

Create `internal/commands/update.go`:

```go
package commands

import (
	"fmt"
	"strings"
)

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
```

- [ ] **Step 4: Run tests — verify they pass**

```bash
go test ./internal/commands/... -run "TestDetermineAssetName|TestNormalizeVersion" -v
```

Expected:
```
--- PASS: TestDetermineAssetName/linux/amd64
--- PASS: TestDetermineAssetName/darwin/amd64
--- PASS: TestDetermineAssetName/darwin/arm64
--- PASS: TestDetermineAssetName/windows/amd64
--- PASS: TestNormalizeVersion/0.2.0
--- PASS: TestNormalizeVersion/v0.2.0
--- PASS: TestNormalizeVersion/1.0.0
--- PASS: TestNormalizeVersion/v1.0.0
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/commands/update.go internal/commands/update_test.go
git commit -m "feat(update): add determineAssetName and normalizeVersion with tests"
```

---

## Task 8: `jira update` — Full Command Implementation

**Files:**
- Modify: `internal/commands/update.go`

- [ ] **Step 1: Replace `update.go` with the full implementation**

Replace the contents of `internal/commands/update.go` with:

```go
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
		fmt.Scanln(&answer)
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
	defer resp.Body.Close()

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
	defer os.Remove(tmpName)

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing download: %w", err)
	}
	tmpFile.Close()

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
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close()

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
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == binaryName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
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
```

- [ ] **Step 2: Run all tests to verify nothing broke**

```bash
go test ./...
```

Expected: all tests pass including the two new ones.

- [ ] **Step 3: Build and smoke-test the command**

```bash
go build -o /tmp/jira-test ./cmd/jira
/tmp/jira-test update --help
```

Expected output includes:
```
Check for a newer version on GitHub Releases and update the binary.
...
Flags:
      --check   Check for updates without downloading (exit 1 if update available)
  -h, --help    help for update
      --path    Override install directory
  -y, --yes     Skip confirmation prompt
```

- [ ] **Step 4: Test `--check` flag with a dev build (should error gracefully)**

```bash
/tmp/jira-test update --check
```

Expected: `Error: cannot update a dev build — install a tagged release first`

- [ ] **Step 5: Commit**

```bash
git add internal/commands/update.go
git commit -m "feat(update): add jira update command with --check, --yes, and --path flags"
```

---

## Task 9: Push and Verify CI

- [ ] **Step 1: Push all commits to origin**

```bash
git push origin main
```

- [ ] **Step 2: Verify CI passes**

Go to `https://github.com/aJesus37/jira-go/actions` and confirm the CI workflow run triggered by the push passes all steps (mod-tidy, fmt, lint, test).

- [ ] **Step 3: Tag and trigger a release (v0.2.0)**

```bash
git tag v0.2.0
git push origin v0.2.0
```

- [ ] **Step 4: Verify the release workflow completes**

Go to `https://github.com/aJesus37/jira-go/actions` and confirm the Release workflow run completes. Then go to `https://github.com/aJesus37/jira-go/releases` and verify:
- Release `v0.2.0` exists
- 4 binary archives are attached: `jira_linux_amd64.tar.gz`, `jira_darwin_amd64.tar.gz`, `jira_darwin_arm64.tar.gz`, `jira_windows_amd64.zip`
- `checksums.txt` is attached
- Release body contains the CHANGELOG content for v0.2.0

- [ ] **Step 5: Test `jira update --check` with the real release**

Download `jira_linux_amd64.tar.gz` from the release, extract it, run:

```bash
./jira update --check
```

Expected (since this is already v0.2.0):
```
Already up to date (v0.2.0)
```

If you have an older build handy, `--check` should print:
```
Current version: v0.1.0
Latest version:  v0.2.0
Update available.
```
and exit with code 1.
