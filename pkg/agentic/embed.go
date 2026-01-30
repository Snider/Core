package agentic

import (
	"embed"
	"strings"
)

//go:embed prompts/*.md
var promptsFS embed.FS

// Prompt returns the content of an embedded prompt file.
// Name should be without the .md extension (e.g., "commit").
func Prompt(name string) string {
	data, err := promptsFS.ReadFile("prompts/" + name + ".md")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
