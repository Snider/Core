#!/bin/bash
# Extract actionable items from core CLI output
# Called PostToolUse on Bash commands that run core

read -r input
COMMAND=$(echo "$input" | jq -r '.tool_input.command // empty')
OUTPUT=$(echo "$input" | jq -r '.tool_output.output // empty')

CONTEXT_SCRIPT="$(dirname "$0")/capture-context.sh"

# Extract actionables from specific core commands
case "$COMMAND" in
    "core go qa"*|"core go test"*|"core go lint"*)
        # Extract error/warning lines
        echo "$OUTPUT" | grep -E "^(ERROR|WARN|FAIL|---)" | head -5 | while read -r line; do
            "$CONTEXT_SCRIPT" "$line" "core go"
        done
        ;;
    "core php test"*|"core php analyse"*)
        # Extract PHP errors
        echo "$OUTPUT" | grep -E "^(FAIL|Error|×)" | head -5 | while read -r line; do
            "$CONTEXT_SCRIPT" "$line" "core php"
        done
        ;;
    "core build"*)
        # Extract build errors
        echo "$OUTPUT" | grep -E "^(error|cannot|undefined)" | head -5 | while read -r line; do
            "$CONTEXT_SCRIPT" "$line" "core build"
        done
        ;;
esac

# Pass through
echo "$input"
