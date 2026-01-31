#!/bin/bash
# Block creation of random .md files - keeps docs consolidated

read -r input
FILE_PATH=$(echo "$input" | jq -r '.tool_input.file_path // empty')

if [[ -n "$FILE_PATH" ]]; then
    # Allow known documentation files
    case "$FILE_PATH" in
        *README.md|*CLAUDE.md|*AGENTS.md|*CONTRIBUTING.md|*CHANGELOG.md|*LICENSE.md)
            echo "$input"
            exit 0
            ;;
        # Allow docs/ directory
        */docs/*.md|*/docs/**/*.md)
            echo "$input"
            exit 0
            ;;
        # Block other .md files
        *.md)
            echo '{"decision": "block", "message": "Use README.md or docs/ for documentation. Random .md files clutter the repo."}'
            exit 0
            ;;
    esac
fi

echo "$input"
