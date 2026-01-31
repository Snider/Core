#!/bin/bash
# Session start: Read scratchpad if recent, otherwise start fresh
# 3 hour window - if older, you've moved on mentally

STATE_FILE="${HOME}/.claude/sessions/scratchpad.md"
THREE_HOURS=10800  # seconds

if [[ -f "$STATE_FILE" ]]; then
    # Get timestamp from file
    FILE_TS=$(grep -E '^timestamp:' "$STATE_FILE" 2>/dev/null | cut -d' ' -f2)
    NOW=$(date '+%s')

    if [[ -n "$FILE_TS" ]]; then
        AGE=$((NOW - FILE_TS))

        if [[ $AGE -lt $THREE_HOURS ]]; then
            # Recent - read it back
            echo "[SessionStart] Found recent scratchpad ($(($AGE / 60)) min ago)" >&2
            echo "[SessionStart] Reading previous state..." >&2
            echo "" >&2
            cat "$STATE_FILE" >&2
            echo "" >&2
        else
            # Stale - delete and start fresh
            rm -f "$STATE_FILE"
            echo "[SessionStart] Previous session >3h old - starting fresh" >&2
        fi
    else
        # No timestamp, delete it
        rm -f "$STATE_FILE"
    fi
fi

exit 0
