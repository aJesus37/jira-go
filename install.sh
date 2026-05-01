#!/usr/bin/env bash
set -euo pipefail

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Fetch latest release tag from GitHub
echo "Fetching latest release..."
LATEST=$(curl -fsSL https://api.github.com/repos/aJesus37/jira-go/releases/latest \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
  echo "Could not determine latest release." >&2
  exit 1
fi

FILENAME="jira_${OS}_${ARCH}.tar.gz"
URL="https://github.com/aJesus37/jira-go/releases/download/${LATEST}/${FILENAME}"

# Download to a temp directory
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $FILENAME ($LATEST)..."
curl -fsSL "$URL" -o "$TMP/jira.tar.gz"
tar -xzf "$TMP/jira.tar.gz" -C "$TMP"

# Install binary to PATH and skills for Claude Code
echo "Installing..."
"$TMP/jira" install --force
"$TMP/jira" skills install

echo ""
echo "Done! Run 'jira init' to configure your Jira credentials."
