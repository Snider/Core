#!/bin/bash
# Install the core skill globally for Claude Code
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/host-uk/core/main/.claude/skills/core/install.sh | bash
#
# Or if you have the repo cloned:
#   ./.claude/skills/core/install.sh

set -e

SKILL_DIR="$HOME/.claude/skills/core"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if running from repo or downloading
if [ -f "$SCRIPT_DIR/SKILL.md" ]; then
    SOURCE_DIR="$SCRIPT_DIR"
else
    # Download from GitHub
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT

    echo "Downloading core skill..."
    curl -fsSL "https://raw.githubusercontent.com/host-uk/core/main/.claude/skills/core/SKILL.md" -o "$TEMP_DIR/SKILL.md"
    SOURCE_DIR="$TEMP_DIR"
fi

# Create skills directory if needed
mkdir -p "$SKILL_DIR"

# Copy skill file
cp "$SOURCE_DIR/SKILL.md" "$SKILL_DIR/SKILL.md"

echo "Installed core skill to $SKILL_DIR"
echo ""
echo "Usage:"
echo "  - Claude will auto-invoke when working in host-uk repos"
echo "  - Or type /core to invoke manually"
echo ""
echo "Commands available: core test, core build, core ci, core work, etc."