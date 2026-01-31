package cli

var glyphMapUnicode = map[string]string{
	":check:": "✓", ":cross:": "✗", ":warn:": "⚠", ":info:": "ℹ",
	":question:": "?", ":skip:": "○", ":dot:": "●", ":circle:": "◯",
	":arrow_right:": "→", ":arrow_left:": "←", ":arrow_up:": "↑", ":arrow_down:": "↓",
	":pointer:": "▶", ":bullet:": "•", ":dash:": "─", ":pipe:": "│",
	":corner:": "└", ":tee:": "├", ":pending:": "…", ":spinner:": "⠋",
}

var glyphMapEmoji = map[string]string{
	":check:": "✅", ":cross:": "❌", ":warn:": "⚠️", ":info:": "ℹ️",
	":question:": "❓", ":skip:": "⏭️", ":dot:": "🔵", ":circle:": "⚪",
	":arrow_right:": "➡️", ":arrow_left:": "⬅️", ":arrow_up:": "⬆️", ":arrow_down:": "⬇️",
	":pointer:": "▶️", ":bullet:": "•", ":dash:": "─", ":pipe:": "│",
	":corner:": "└", ":tee:": "├", ":pending:": "⏳", ":spinner:": "🔄",
}

var glyphMapASCII = map[string]string{
	":check:": "[OK]", ":cross:": "[FAIL]", ":warn:": "[WARN]", ":info:": "[INFO]",
	":question:": "[?]", ":skip:": "[SKIP]", ":dot:": "[*]", ":circle:": "[ ]",
	":arrow_right:": "->", ":arrow_left:": "<-", ":arrow_up:": "^", ":arrow_down:": "v",
	":pointer:": ">", ":bullet:": "*", ":dash:": "-", ":pipe:": "|",
	":corner:": "`", ":tee:": "+", ":pending:": "...", ":spinner:": "-",
}
