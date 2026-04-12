#!/bin/bash
set -e

AGENTS_DIR="$HOME/.agents/skills/jira-go"
CLAUDE_SKILLS_DIR="$HOME/.claude/skills"
SKILLS_SOURCE="$(cd "$(dirname "$0")" && pwd)/skills"

echo "Installing jira-go skills to ~/.agents/skills/jira-go/"

# Create parent dir if needed
mkdir -p "$(dirname "$AGENTS_DIR")"

# Remove existing installation
rm -rf "$AGENTS_DIR"
mkdir -p "$AGENTS_DIR"

# Copy skills
cp -r "$SKILLS_SOURCE"/* "$AGENTS_DIR/"

echo "Copied skills to $AGENTS_DIR"

# Create symlinks in ~/.claude/skills/
for skill_dir in "$AGENTS_DIR"/*/; do
    skill_name=$(basename "$skill_dir")
    link_path="$CLAUDE_SKILLS_DIR/$skill_name"

    # Remove existing symlink or file
    rm -rf "$link_path"

    # Create symlink
    ln -sfn "$skill_dir" "$link_path"
    echo "Linked $link_path"
done

echo ""
echo "Installed skills:"
ls -la "$AGENTS_DIR"
