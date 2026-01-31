#!/bin/bash
# Warn about debug statements left in code after edits

read -r input
FILE_PATH=$(echo "$input" | jq -r '.tool_input.file_path // empty')

if [[ -n "$FILE_PATH" && -f "$FILE_PATH" ]]; then
    case "$FILE_PATH" in
        *.go)
            # Check for fmt.Println, log.Println debug statements
            if grep -n "fmt\.Println\|log\.Println" "$FILE_PATH" 2>/dev/null | head -3 | grep -q .; then
                echo "[Hook] WARNING: Debug prints found in $FILE_PATH" >&2
                grep -n "fmt\.Println\|log\.Println" "$FILE_PATH" 2>/dev/null | head -3 >&2
            fi
            ;;
        *.php)
            # Check for dd(), dump(), var_dump(), print_r()
            if grep -n "dd(\|dump(\|var_dump(\|print_r(" "$FILE_PATH" 2>/dev/null | head -3 | grep -q .; then
                echo "[Hook] WARNING: Debug statements found in $FILE_PATH" >&2
                grep -n "dd(\|dump(\|var_dump(\|print_r(" "$FILE_PATH" 2>/dev/null | head -3 >&2
            fi
            ;;
    esac
fi

# Pass through the input
echo "$input"
