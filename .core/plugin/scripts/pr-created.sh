#!/bin/bash
# Log PR URL and provide review command after PR creation

read -r input
COMMAND=$(echo "$input" | jq -r '.tool_input.command // empty')
OUTPUT=$(echo "$input" | jq -r '.tool_output.output // empty')

if [[ "$COMMAND" == *"gh pr create"* ]]; then
    PR_URL=$(echo "$OUTPUT" | grep -oE 'https://github.com/[^/]+/[^/]+/pull/[0-9]+' | head -1)
    if [[ -n "$PR_URL" ]]; then
        REPO=$(echo "$PR_URL" | sed -E 's|https://github.com/([^/]+/[^/]+)/pull/[0-9]+|\1|')
        PR_NUM=$(echo "$PR_URL" | sed -E 's|.*/pull/([0-9]+)|\1|')
        echo "[Hook] PR created: $PR_URL" >&2
        echo "[Hook] To review: gh pr review $PR_NUM --repo $REPO" >&2
    fi
fi

echo "$input"
