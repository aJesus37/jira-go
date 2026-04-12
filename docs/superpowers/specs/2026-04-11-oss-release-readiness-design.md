# OSS Release Readiness Design

**Date:** 2026-04-11  
**Status:** Approved  

## Overview

jira-go is feature-complete and needs the scaffolding to be a proper open source project: a corrected module path and README, a CI quality gate, a GoReleaser-based cross-platform release pipeline, and a `jira update` self-update command.

---

## Section 1 — Module Path & Stale Reference Cleanup

### Module path

Replace all occurrences of the placeholder `github.com/user/jira-go` and `github.com/user/jira` with `github.com/aJesus37/jira-go`. Affected files:

- `go.mod` (module declaration)
- All `*.go` files with import paths
- `README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`

### Go version

Fix `go 1.25.7` in `go.mod` to `go 1.21` (matching what CONTRIBUTING.md documents; 1.25.7 is a non-existent pre-release).

### README corrections

- **Remove** `--no-cache` and `--cache-ttl` from the Global Flags section (these flags were deleted)
- **Add** owner management keyboard shortcut (`o` — manage owners) to the Sprint Board keyboard shortcuts table
- **Fix** installation instructions to reference `github.com/aJesus37/jira-go`
- **Fix** the Cache Management section — verify `jira cache status/clear/path` commands still exist; remove if not

---

## Section 2 — CI Pipeline

**File:** `.github/workflows/ci.yml`

**Triggers:** `push` to `main`, `pull_request` targeting `main`

**Steps:**
1. `actions/checkout@v4`
2. `actions/setup-go@v5` with version from `go.mod`
3. `actions/cache@v4` for Go module cache
4. `go mod tidy` + `git diff --exit-code` (catches un-tidied modules)
5. `go fmt ./...` + `git diff --exit-code` (catches unformatted code)
6. `golangci-lint/golangci-lint-action@v6`
7. `go test ./...`

CI is a quality gate only — no build artifacts produced.

---

## Section 3 — GoReleaser + Release Pipeline

### `.goreleaser.yaml`

- **Targets:** `linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- **ldflags:** inject `version`, `commit`, `buildDate` (same vars used in `Taskfile.yml`)
- **Archives:** `.tar.gz` for Linux/macOS, `.zip` for Windows
- **Naming:** `jira_{{ .Os }}_{{ .Arch }}` for predictable filenames
- **Checksums:** `checksums.txt` (sha256)
- **Release body:** auto-extracted from `CHANGELOG.md` current version section

### `.github/workflows/release.yml`

**Trigger:** `push` with tag matching `v*`

**Steps:**
1. `actions/checkout@v4` with `fetch-depth: 0` (GoReleaser needs full history)
2. `actions/setup-go@v5`
3. `goreleaser/goreleaser-action@v6` with `GITHUB_TOKEN`

**Release flow for maintainer:**
```bash
git tag v0.2.0
git push --tags
```
GoReleaser handles everything else.

### CHANGELOG update

- Version `[Unreleased]` → `[0.2.0] - 2026-04-11`
- Add new empty `[Unreleased]` section above it
- Update comparison URLs at the bottom to point to `aJesus37/jira-go`

---

## Section 4 — `jira update` Command

**File:** `internal/commands/update.go`

### Behavior

**`jira update`** (default):
1. Call GitHub Releases API: `GET https://api.github.com/repos/aJesus37/jira-go/releases/latest`
2. Compare `tag_name` against embedded `version` var
3. If already latest → print "Already up to date (vX.Y.Z)" and exit 0
4. If newer available → print current vs latest, prompt "Update? [y/N]"
5. On confirm:
   - Determine asset name for current OS/arch (e.g. `jira_linux_amd64.tar.gz`)
   - Download asset to a temp file
   - Extract binary from archive
   - Write to user-local install directory via `getInstallDirectory()` (reused from `install.go`) — no sudo required
   - Print "Updated to vX.Y.Z — restart your shell if needed"

### Flags

| Flag | Behavior |
|------|----------|
| `--check` | Print version comparison only, no download. Exit code 1 if update available (useful in scripts/CI). |
| `--yes` / `-y` | Skip confirmation prompt, update immediately. |
| `--path` | Override destination directory (same as `install --path`). |

### Install path

Always writes to the user-local directory returned by `getInstallDirectory()`:
- **Linux:** `~/.local/bin`
- **macOS:** `~/.local/bin` (fallback: `~/bin`)
- **Windows:** `%LOCALAPPDATA%\Programs\jira`

No sudo required. If the user originally installed globally, the updated binary lands in the user-local path, which typically takes precedence in `PATH`.

`getInstallDirectory()` is already implemented in `install.go` — `update.go` reuses it directly.

### Error handling

| Scenario | Behavior |
|----------|----------|
| No network / API unavailable | Clear error: "Could not reach GitHub API: \<err\>" |
| `version` is `"dev"` (built from source) | Warn: "Cannot update a dev build" and exit 1 |
| Asset not found for current OS/arch | Error: "No release asset found for \<os\>/\<arch\>" |
| Download fails mid-stream | Temp file cleaned up, error surfaced |

### Asset naming convention (must match GoReleaser config)

| OS/Arch | Asset filename |
|---------|---------------|
| linux/amd64 | `jira_linux_amd64.tar.gz` |
| darwin/amd64 | `jira_darwin_amd64.tar.gz` |
| darwin/arm64 | `jira_darwin_arm64.tar.gz` |
| windows/amd64 | `jira_windows_amd64.zip` |

---

## Out of Scope

- Homebrew tap (can be added after first release)
- Linux ARM64 builds (can be added to GoReleaser targets later)
- Signed/notarized binaries (macOS Gatekeeper)
- Docker image

---

## Implementation Order

1. Module path + README/CHANGELOG cleanup (unblocks everything else)
2. CI workflow (quick win, catches regressions immediately)
3. GoReleaser config + release workflow (enables actual releases)
4. `jira update` command (depends on GoReleaser asset naming being finalized)
