#!/bin/bash
# Pre-compact: Save minimal state for Claude to resume after auto-compact
#
# Captures:
# - Working directory + branch
# - Git status (files touched)
# - Todo state (in_progress items)
# - Context facts (decisions, actionables)

STATE_FILE="${HOME}/.claude/sessions/scratchpad.md"
CONTEXT_FILE="${HOME}/.claude/sessions/context.json"
TIMESTAMP=$(date '+%s')
CWD=$(pwd)

mkdir -p "${HOME}/.claude/sessions"

# Get todo state
TODOS=""
if [[ -f "${HOME}/.claude/todos/current.json" ]]; then
    TODOS=$(cat "${HOME}/.claude/todos/current.json" 2>/dev/null | head -50)
fi

# Get git status
GIT_STATUS=""
BRANCH=""
if git rev-parse --git-dir > /dev/null 2>&1; then
    GIT_STATUS=$(git status --short 2>/dev/null | head -15)
    BRANCH=$(git branch --show-current 2>/dev/null)
fi

# Get context facts
CONTEXT=""
if [[ -f "$CONTEXT_FILE" ]]; then
    CONTEXT=$(jq -r '.[] | "- [\(.source)] \(.fact)"' "$CONTEXT_FILE" 2>/dev/null | tail -10)
fi

cat > "$STATE_FILE" << EOF
---
timestamp: ${TIMESTAMP}
cwd: ${CWD}
branch: ${BRANCH:-none}
---

# Resume After Compact

You were mid-task. Do NOT assume work is complete.

## Project
\`${CWD}\` on \`${BRANCH:-no branch}\`

## Files Changed
\`\`\`
${GIT_STATUS:-none}
\`\`\`

## Todos (in_progress = NOT done)
\`\`\`json
${TODOS:-check /todos}
\`\`\`

## Context (decisions & actionables)
${CONTEXT:-none captured}

## Next
Continue the in_progress todo.
EOF

echo "[PreCompact] Snapshot saved" >&2
exit 0
