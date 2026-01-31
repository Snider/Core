#!/bin/bash
# PreToolUse hook: Block dangerous commands, enforce core CLI
#
# BLOCKS:
# - Raw go commands (use core go *)
# - Destructive grep patterns (sed -i, xargs rm, etc.)
# - Mass file operations (rm -rf, mv/cp with wildcards)
# - Any sed outside of safe patterns
#
# This prevents "efficient shortcuts" that nuke codebases

read -r input
command=$(echo "$input" | jq -r '.tool_input.command // empty')

# === HARD BLOCKS - Never allow these ===

# Block rm -rf, rm -r (except for known safe paths like node_modules, vendor, .cache)
if echo "$command" | grep -qE 'rm\s+(-[a-zA-Z]*r[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r|--recursive)'; then
    # Allow only specific safe directories
    if ! echo "$command" | grep -qE 'rm\s+(-rf|-r)\s+(node_modules|vendor|\.cache|dist|build|__pycache__|\.pytest_cache|/tmp/)'; then
        echo '{"decision": "block", "message": "BLOCKED: Recursive delete is not allowed. Delete files individually or ask the user to run this command."}'
        exit 0
    fi
fi

# Block mv/cp with wildcards (mass file moves)
if echo "$command" | grep -qE '(mv|cp)\s+.*\*'; then
    echo '{"decision": "block", "message": "BLOCKED: Mass file move/copy with wildcards is not allowed. Move files individually."}'
    exit 0
fi

# Block xargs with rm, mv, cp (mass operations)
if echo "$command" | grep -qE 'xargs\s+.*(rm|mv|cp)'; then
    echo '{"decision": "block", "message": "BLOCKED: xargs with file operations is not allowed. Too risky for mass changes."}'
    exit 0
fi

# Block find -exec with rm, mv, cp
if echo "$command" | grep -qE 'find\s+.*-exec\s+.*(rm|mv|cp)'; then
    echo '{"decision": "block", "message": "BLOCKED: find -exec with file operations is not allowed. Too risky for mass changes."}'
    exit 0
fi

# Block ALL sed -i (in-place editing)
if echo "$command" | grep -qE 'sed\s+(-[a-zA-Z]*i|--in-place)'; then
    echo '{"decision": "block", "message": "BLOCKED: sed -i (in-place edit) is never allowed. Use the Edit tool for file changes."}'
    exit 0
fi

# Block sed piped to file operations
if echo "$command" | grep -qE 'sed.*\|.*tee|sed.*>'; then
    echo '{"decision": "block", "message": "BLOCKED: sed with file output is not allowed. Use the Edit tool for file changes."}'
    exit 0
fi

# Block grep with -l piped to xargs/rm/sed (the classic codebase nuke pattern)
if echo "$command" | grep -qE 'grep\s+.*-l.*\|'; then
    echo '{"decision": "block", "message": "BLOCKED: grep -l piped to other commands is the classic codebase nuke pattern. Not allowed."}'
    exit 0
fi

# Block perl -i, awk with file redirection (sed alternatives)
if echo "$command" | grep -qE 'perl\s+-[a-zA-Z]*i|awk.*>'; then
    echo '{"decision": "block", "message": "BLOCKED: In-place file editing with perl/awk is not allowed. Use the Edit tool."}'
    exit 0
fi

# === REQUIRE CORE CLI ===

# Block raw go commands
case "$command" in
    "go test"*|"go build"*|"go fmt"*|"go mod tidy"*|"go vet"*|"go run"*)
        echo '{"decision": "block", "message": "Use `core go test`, `core build`, `core go fmt --fix`, etc. Raw go commands are not allowed."}'
        exit 0
        ;;
    "go "*)
        # Other go commands - warn but allow
        echo '{"decision": "block", "message": "Prefer `core go *` commands. If core does not have this command, ask the user."}'
        exit 0
        ;;
esac

# Block raw php commands
case "$command" in
    "php artisan serve"*|"./vendor/bin/pest"*|"./vendor/bin/pint"*|"./vendor/bin/phpstan"*)
        echo '{"decision": "block", "message": "Use `core php dev`, `core php test`, `core php fmt`, `core php analyse`. Raw php commands are not allowed."}'
        exit 0
        ;;
    "composer test"*|"composer lint"*)
        echo '{"decision": "block", "message": "Use `core php test` or `core php fmt`. Raw composer commands are not allowed."}'
        exit 0
        ;;
esac

# Block golangci-lint directly
if echo "$command" | grep -qE '^golangci-lint'; then
    echo '{"decision": "block", "message": "Use `core go lint` instead of golangci-lint directly."}'
    exit 0
fi

# === APPROVED ===
echo '{"decision": "approve"}'
