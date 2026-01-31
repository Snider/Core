#!/bin/bash
# Post-commit hook: Check for uncommitted work that might get lost
#
# After committing task-specific files, check if there's other work
# in the repo that should be committed or stashed

read -r input
COMMAND=$(echo "$input" | jq -r '.tool_input.command // empty')

# Only run after git commit
if ! echo "$COMMAND" | grep -qE '^git commit'; then
    echo "$input"
    exit 0
fi

# Check for remaining uncommitted changes
UNSTAGED=$(git diff --name-only 2>/dev/null | wc -l | tr -d ' ')
STAGED=$(git diff --cached --name-only 2>/dev/null | wc -l | tr -d ' ')
UNTRACKED=$(git ls-files --others --exclude-standard 2>/dev/null | wc -l | tr -d ' ')

TOTAL=$((UNSTAGED + STAGED + UNTRACKED))

if [[ $TOTAL -gt 0 ]]; then
    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "[PostCommit] WARNING: Uncommitted work remains" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2

    if [[ $UNSTAGED -gt 0 ]]; then
        echo "  Modified (unstaged): $UNSTAGED files" >&2
        git diff --name-only 2>/dev/null | head -5 | sed 's/^/    /' >&2
        [[ $UNSTAGED -gt 5 ]] && echo "    ... and $((UNSTAGED - 5)) more" >&2
    fi

    if [[ $STAGED -gt 0 ]]; then
        echo "  Staged (not committed): $STAGED files" >&2
        git diff --cached --name-only 2>/dev/null | head -5 | sed 's/^/    /' >&2
    fi

    if [[ $UNTRACKED -gt 0 ]]; then
        echo "  Untracked: $UNTRACKED files" >&2
        git ls-files --others --exclude-standard 2>/dev/null | head -5 | sed 's/^/    /' >&2
        [[ $UNTRACKED -gt 5 ]] && echo "    ... and $((UNTRACKED - 5)) more" >&2
    fi

    echo "" >&2
    echo "Consider: commit these, stash them, or confirm they're intentionally left" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
fi

echo "$input"
