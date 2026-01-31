#!/bin/bash
# Capture context facts from tool output or conversation
# Called by PostToolUse hooks to extract actionable items
#
# Stores in ~/.claude/sessions/context.json as:
# [{"fact": "...", "source": "core go qa", "ts": 1234567890}, ...]

CONTEXT_FILE="${HOME}/.claude/sessions/context.json"
TIMESTAMP=$(date '+%s')
THREE_HOURS=10800

mkdir -p "${HOME}/.claude/sessions"

# Initialize if missing or stale
if [[ -f "$CONTEXT_FILE" ]]; then
    FIRST_TS=$(jq -r '.[0].ts // 0' "$CONTEXT_FILE" 2>/dev/null)
    NOW=$(date '+%s')
    AGE=$((NOW - FIRST_TS))
    if [[ $AGE -gt $THREE_HOURS ]]; then
        echo "[]" > "$CONTEXT_FILE"
    fi
else
    echo "[]" > "$CONTEXT_FILE"
fi

# Read input (fact and source passed as args or stdin)
FACT="${1:-}"
SOURCE="${2:-manual}"

if [[ -z "$FACT" ]]; then
    # Try reading from stdin
    read -r FACT
fi

if [[ -n "$FACT" ]]; then
    # Append to context (keep last 20 items)
    jq --arg fact "$FACT" --arg source "$SOURCE" --argjson ts "$TIMESTAMP" \
        '. + [{"fact": $fact, "source": $source, "ts": $ts}] | .[-20:]' \
        "$CONTEXT_FILE" > "${CONTEXT_FILE}.tmp" && mv "${CONTEXT_FILE}.tmp" "$CONTEXT_FILE"

    echo "[Context] Saved: $FACT" >&2
fi

exit 0
