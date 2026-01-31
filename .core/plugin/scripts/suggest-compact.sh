#!/bin/bash
# Suggest /compact at logical intervals to manage context window
# Tracks tool calls per session, suggests compaction every 50 calls

SESSION_ID="${CLAUDE_SESSION_ID:-$$}"
COUNTER_FILE="/tmp/claude-tool-count-${SESSION_ID}"
THRESHOLD="${COMPACT_THRESHOLD:-50}"

# Read or initialize counter
if [[ -f "$COUNTER_FILE" ]]; then
    COUNT=$(($(cat "$COUNTER_FILE") + 1))
else
    COUNT=1
fi

echo "$COUNT" > "$COUNTER_FILE"

# Suggest compact at threshold
if [[ $COUNT -eq $THRESHOLD ]]; then
    echo "[Compact] ${THRESHOLD} tool calls - consider /compact if transitioning phases" >&2
fi

# Suggest at intervals after threshold
if [[ $COUNT -gt $THRESHOLD ]] && [[ $((COUNT % 25)) -eq 0 ]]; then
    echo "[Compact] ${COUNT} tool calls - good checkpoint for /compact" >&2
fi

exit 0
